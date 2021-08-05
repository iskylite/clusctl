package service

import (
	"context"
	"fmt"
	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"
	"strconv"
	"time"

	"google.golang.org/grpc"
)

type RunCmdClientService struct {
	cmd      string
	nodelist string
	port     string
	client   pb.RpcServiceClient
	ctx      context.Context
}

func NewRunCmdClientService(ctx context.Context, cmd, nodelist, port string) (*RunCmdClientService, error) {
	nodes := utils.ExpNodes(nodelist)
	return newRunCmdClientService(ctx, cmd, port, nodes)
}

func newRunCmdClientService(ctx context.Context, cmd, port string, nodes []string) (*RunCmdClientService, error) {
	batchNode := nodes[0]
	allocNodelist := ""
	if len(nodes) > 1 {
		allocNodelist = utils.ConvertNodelist(nodes[1:])
	}
	addr := fmt.Sprintf("%s:%s", batchNode, port)
	// grpc 最大传输数据大小 每个子节点传输3M大小*总节点数
	grpcOptions := grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024 * 1024 * 3 * len(nodes)))
	ctx, _ = context.WithTimeout(ctx, time.Second)
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock(), grpcOptions)
	if err != nil {
		return nil, utils.GrpcErrorWrapper(err)
	}
	log.Debugf("Dial Server %s\n", addr)
	client := pb.NewRpcServiceClient(conn)
	log.Debugf("Connect Server %s\n", addr)
	return &RunCmdClientService{
		cmd:      cmd,
		client:   client,
		nodelist: allocNodelist,
		ctx:      ctx,
		port:     port,
	}, err
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
	client, err := NewRunCmdClientService(ctx, cmd, nodes, strconv.Itoa(port))
	if err != nil {
		log.Error(err)
		return
	}
	replays, err := client.RunCmd(int32(width))
	if err != nil {
		log.Error(err)
		return
	}
	client.Gather(replays, nodes, list)
}
