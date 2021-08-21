package service

import (
	"context"
	"errors"
	"fmt"
	"myclush/logger"
	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type RunCmdClientService struct {
	cmd      string
	nodelist string
	port     string
	node     string
	conn     *grpc.ClientConn
	client   pb.RpcServiceClient
	ctx      context.Context
}

func NewRunCmdClientService(ctx context.Context, cmd, nodelist, port string) (*RunCmdClientService, []string, error) {
	nodes := utils.ExpNodes(nodelist)
	return newRunCmdClientService(ctx, cmd, port, nodes)
}

func newRunCmdClientService(ctx context.Context, cmd, port string, nodes []string) (*RunCmdClientService, []string, error) {
	nodesNum := len(nodes)
	down := make([]string, 0)
	var conn *grpc.ClientConn
	var stream pb.RpcServiceClient
	var err error
	p := new(RunCmdClientService)
	p.cmd, p.port, p.ctx = cmd, port, ctx
	grpcOptions := grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024 * 1024 * 3 * len(nodes)))
	for i := 0; i < nodesNum; i++ {
		node := nodes[i]
		conn, stream, err = checkConn(ctx, node, port, grpcOptions)
		if err != nil {
			down = append(down, node)
			continue
		}
		p.node, p.conn, p.client = node, conn, stream
		p.nodelist = utils.ConvertNodelist(nodes[i+1 : nodesNum])
		break
	}
	// 只要有一个连接成功，那么err就会被赋值为nil，否则则是连接失败的错误
	// 故当err为错误的时候，所有节点都连接失败
	return p, down, err
}

func checkConn(ctx context.Context, node, port string, grpcOptions grpc.DialOption) (*grpc.ClientConn, pb.RpcServiceClient, error) {
	addr := fmt.Sprintf("%s:%s", node, port)
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpcOptions)
	if err != nil {
		logger.Error(err)
		return nil, nil, err
	}
	client := pb.NewRpcServiceClient(conn)
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			logger.Debugf("connect timeout for %s\n", node)
			if conn != nil {
				conn.Close()
			}
			return nil, nil, errors.New("timeout")
		default:
			if conn.GetState() == connectivity.Ready {
				log.Debugf("Gen client stream -> %s\n", addr)
				return conn, client, utils.GrpcErrorWrapper(err)
			}
		}
	}
}

func (r *RunCmdClientService) RunCmd(width int32) ([]*pb.Replay, error) {
	req := &pb.CmdReq{Cmd: r.cmd, Nodelist: r.nodelist, Width: width, Port: r.port}
	log.Debugf("Command %s Start ...\n", r.cmd)
	resp, err := r.client.RunCmd(r.ctx, req)
	if err != nil {
		return nil, utils.GrpcErrorWrapper(err)
	}
	log.Debugf("Command %s End\n", r.cmd)
	resps := make([]*pb.Replay, 0)
	if resp != nil {
		resps = append(resps, resp.Replay...)
	}
	return resps, nil
}

func (r *RunCmdClientService) CloseConn() {
	r.conn.Close()
	logger.Debugf("close conn %s\n", r.node)
}

func (r *RunCmdClientService) Gather(replay []*pb.Replay, nodelist string, flag bool) {
	if flag {
		// 顺序打印结果
		nodes := utils.ExpNodes(nodelist)
		replayMap := make(map[string]*pb.Replay)
		for _, rep := range replay {
			nodelist := utils.ExpNodes(rep.Nodelist)
			for _, node := range nodelist {
				replayMap[node] = rep
			}
		}
		for _, node := range nodes {
			rep, ok := replayMap[node]
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
		gather(replay)
	}
}

func RunCmdClientServiceSetup(ctx context.Context, cancel context.CancelFunc, cmd, nodes string, width, port int, list bool) {
	defer cancel()
	resps := make([]*pb.Replay, 0)
	client, down, err := NewRunCmdClientService(ctx, cmd, nodes, strconv.Itoa(port))
	if err != nil {
		log.Error(err)
		return
	}
	defer client.CloseConn()
	replays, err := client.RunCmd(int32(width))
	if err != nil {
		log.Error(err)
		return
	}
	if len(down) > 0 {
		resps = append(resps, newReplay(false, "connect failed", utils.ConvertNodelist(down)))
	}
	resps = append(resps, replays...)
	client.Gather(resps, nodes, list)
}
