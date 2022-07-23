package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"myclush/global"
	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"
	"os"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/status"
)

type RunCmdClientService struct {
	cmd            string
	nodelist       string
	port           string
	node           string
	num            int
	conn           *grpc.ClientConn
	stream         pb.RpcService_RunCmdClient
	ctx            context.Context
	repliesChannel chan *pb.Reply
}

func NewRunCmdClientService(ctx context.Context, cmd, nodelist, port string, width int32, daemon bool) (*RunCmdClientService, []string, error) {
	nodes := utils.ExpNodes(nodelist)
	return newRunCmdClientService(ctx, cmd, port, nodes, width, global.Authority, daemon)
}

func newRunCmdClientService(ctx context.Context, cmd, port string, nodes []string, width int32, authority grpc.DialOption, daemon bool) (*RunCmdClientService, []string, error) {
	nodesNum := len(nodes)
	down := make([]string, 0)
	var conn *grpc.ClientConn
	var stream pb.RpcService_RunCmdClient
	var err error
	p := new(RunCmdClientService)
	p.cmd, p.port, p.ctx, p.num = cmd, port, ctx, nodesNum
	grpcOptions := grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024 * 1024 * 3 * len(nodes)))
	for i := 0; i < nodesNum; i++ {
		node := nodes[i]
		nodelist := utils.Merge(nodes[i+1 : nodesNum]...)
		log.Debugf("node [%s] -> [%s]\n", node, nodelist)
		req := &pb.CmdReq{Cmd: cmd, Node: node, Nodelist: nodelist, Port: port, Width: width, Daemon: daemon}
		conn, stream, err = checkConn(ctx, node, req, grpcOptions, authority)
		if err != nil {
			down = append(down, node)
			continue
		}
		p.node, p.conn, p.stream = node, conn, stream
		p.nodelist = nodelist
		break
	}
	// 只要有一个连接成功，那么err就会被赋值为nil，否则则是连接失败的错误
	// 故当err为错误的时候，所有节点都连接失败
	return p, down, err
}

func checkConn(ctx context.Context, node string, req *pb.CmdReq, grpcOptions, authority grpc.DialOption) (*grpc.ClientConn, pb.RpcService_RunCmdClient, error) {
	addr := fmt.Sprintf("%s:%s", node, req.Port)
	// dial
	conn, err := grpc.DialContext(ctx, addr, grpcOptions, authority, global.ClientTransportCredentials)
	if err != nil {
		log.Error(err)
		return nil, nil, err
	}
	// runcmd
	client := pb.NewRpcServiceClient(conn)
	stream, err := client.RunCmd(ctx, req)
	if err != nil {
		if conn != nil {
			conn.Close()
		}
		return conn, nil, err
	}
	// waiting
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			log.Errorf("connect timeout for %s\n", node)
			if conn != nil {
				conn.Close()
			}
			return nil, nil, status.Error(codes.DeadlineExceeded, "connect timeout")
		default:
			if conn.GetState() == connectivity.Ready {
				log.Debugf("Gen client stream -> %s\n", addr)
				return conn, stream, err
			}
		}
	}
}

func (r *RunCmdClientService) DiscribeRepliesChannel(repliesChannel chan *pb.Reply) {
	r.repliesChannel = repliesChannel
}

func (r *RunCmdClientService) RunCmd() {
	for {
		select {
		case <-r.ctx.Done():
			log.Debug("get cancel signal")
			return
		default:
			reply, err := r.stream.Recv()
			switch err {
			case nil:
				r.repliesChannel <- reply
			case io.EOF:
				log.Debugf("%s reply io.EOF\n", r.node)
				return
			default:
				log.Errorf("replay err: %v\n", utils.GrpcErrorMsg(err))
				return
			}
		}
	}
}

func (r *RunCmdClientService) CloseConn() {
	r.conn.Close()
	log.Debugf("close conn %s\n", r.node)
}

func (r *RunCmdClientService) Gather(Reply []*pb.Reply, nodelist string, flag bool) {
	if flag {
		// 顺序打印结果
		nodes := utils.ExpNodes(nodelist)
		ReplyMap := make(map[string]*pb.Reply)
		for _, rep := range Reply {
			nodelist := utils.ExpNodes(rep.Nodelist)
			for _, node := range nodelist {
				ReplyMap[node] = rep
			}
		}
		for _, node := range nodes {
			rep, ok := ReplyMap[node]
			if !ok {
				log.ColorWrapperInfo(log.Failed, utils.ExpNodes(node), "hostname unmatched")
				continue
			}
			if rep.Pass {
				log.ColorWrapperInfo(log.Success, utils.ExpNodes(node), rep.Msg)
			} else {
				log.ColorWrapperInfo(log.Failed, utils.ExpNodes(node), rep.Msg)
			}
		}
	} else {
		gather(Reply)
	}
}

func RunCmdClientServiceSetup(ctx context.Context, cancel context.CancelFunc, cmd, nodes string, width, port int, list, daemon bool) {
	defer cancel()
	// establish conn and call remote cmd
	client, down, err := NewRunCmdClientService(ctx, cmd, nodes, strconv.Itoa(port), int32(width), daemon)
	if err != nil {
		log.Error(err)
		return
	}
	defer client.CloseConn()

	// resultSet存储每个节点的是否有响应，用于在enter时输出当前没有拿到响应的节点列表
	resOriginMap, err := hashNodesMap(nodes)
	if err != nil {
		log.Error(err)
		return
	}
	// 获取运行状态下未获取到响应的节点
	go func() {
		for {
			idle := make([]string, 0)
			stdinBuf := bufio.NewReaderSize(os.Stdin, 1)
			key, _ := stdinBuf.ReadByte()
			if key == 10 {
				// press enter
				resOriginMap.Range(func(key, value interface{}) bool {
					if !value.(bool) {
						idle = append(idle, key.(string))
					}
					return true
				})
				idleNodes := utils.Merge(idle...)
				log.Infof("\r等待结果: %s\n", idleNodes)
			}
		}
	}()
	// reply handle
	resps := make([]*pb.Reply, 0)
	repliesChannel := make(chan *pb.Reply)
	var waitc sync.WaitGroup
	waitc.Add(1)
	cnt := 0
	go func() {
		defer waitc.Done()
		fmt.Printf("\r结果汇总: %d/%d", 0, client.num)
		for data := range repliesChannel {
			for _, node := range utils.ExpNodes(data.Nodelist) {
				log.Debugf("接收响应: pass -> %t, node -> %s, msg -> %s\n", data.Pass, node, data.Msg)
				if value, ok := resOriginMap.Load(node); ok {
					// 节点存在于resOriginMap
					if !value.(bool) {
						// 之前未接收到该节点的响应
						resOriginMap.Store(node, true)
						resps = append(resps, newReply(data.Pass, data.Msg, node))
						cnt++
						fmt.Printf("\r结果汇总: %d/%d", cnt, client.num)
					}
				} else {
					// 节点被确认不存在于resOriginMap
					resps = append(resps, newReply(data.Pass, data.Msg, node))
					cnt++
					fmt.Printf("\r结果汇总: %d/%d", cnt, client.num)
				}
			}
		}
		fmt.Printf("\r结果汇总: %d/%d %s\n", cnt, client.num, log.ColorWrapper("EOF", log.Success))
	}()
	// client reply
	for _, node := range down {
		repliesChannel <- newReply(false, "connect or call failed", node)
	}
	client.DiscribeRepliesChannel(repliesChannel)
	client.RunCmd()
	close(repliesChannel)
	waitc.Wait()
	client.Gather(resps, nodes, list)
}
