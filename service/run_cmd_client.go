package service

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"myclush/global"
	"myclush/logger"
	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"
	"os"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
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

func NewRunCmdClientService(ctx context.Context, cmd, nodelist, port string, width int32) (*RunCmdClientService, []string, error) {
	nodes := utils.ExpNodes(nodelist)
	return newRunCmdClientService(ctx, cmd, port, nodes, width, global.Authority)
}

func newRunCmdClientService(ctx context.Context, cmd, port string, nodes []string, width int32, authority grpc.DialOption) (*RunCmdClientService, []string, error) {
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
		req := &pb.CmdReq{Cmd: cmd, Nodelist: nodelist, Port: port, Width: width}
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
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpcOptions, authority)
	if err != nil {
		logger.Error(err)
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
			logger.Debugf("connect timeout for %s\n", node)
			if conn != nil {
				conn.Close()
			}
			return nil, nil, errors.New("connect timeout")
		default:
			if conn.GetState() == connectivity.Ready {
				log.Debugf("Gen client stream -> %s\n", addr)
				return conn, stream, utils.GrpcErrorWrapper(err)
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
	logger.Debugf("close conn %s\n", r.node)
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

func RunCmdClientServiceSetup(ctx context.Context, cancel context.CancelFunc, cmd, nodes string, width, port int, list bool) {
	defer cancel()
	// establish conn and call remote cmd
	client, down, err := NewRunCmdClientService(ctx, cmd, nodes, strconv.Itoa(port), int32(width))
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
				log.Infof("\rProgress Replies: %s\n", idleNodes)
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
		for reply := range repliesChannel {
			fmt.Printf("\rResponse  Replies: %d/%d", cnt, client.num)
			resps = append(resps, reply)
			resOriginMap.Store(reply.Nodelist, true)
			cnt++
		}
		fmt.Printf("\rResponse  Replies: %d/%d %s\n", cnt, client.num, log.ColorWrapper("EOF", log.Success))
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
