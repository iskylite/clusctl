package service

import (
	"context"
	"fmt"
	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"

	"google.golang.org/grpc"
)

type RunCmdClientService struct {
	cmd      string
	nodelist string
	timeout  int32
	port     string
	client   pb.RpcServiceClient
	ctx      context.Context
}

func NewRunCmdClientService(ctx context.Context, cmd, nodelist, port string, timeout int32) (*RunCmdClientService, error) {
	nodes := utils.ExpNodes(nodelist)
	return newRunCmdClientService(ctx, cmd, port, nodes, timeout)
}

func newRunCmdClientService(ctx context.Context, cmd, port string, nodes []string, timeout int32) (*RunCmdClientService, error) {
	batchNode := nodes[0]
	allocNodelist := ""
	if len(nodes) > 1 {
		allocNodelist = utils.ConvertNodelist(nodes[1:])
	}
	addr := fmt.Sprintf("%s:%s", batchNode, port)
	grpcOptions := grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024 * 1024 * 3 * len(nodes)))
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpcOptions)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	log.Debugf("Dial Server [%s]\n", addr)
	client := pb.NewRpcServiceClient(conn)
	log.Debugf("Connect Server [%s]\n", addr)
	return &RunCmdClientService{
		cmd:      cmd,
		timeout:  timeout,
		client:   client,
		nodelist: allocNodelist,
		ctx:      ctx,
		port:     port,
	}, err
}

func (r *RunCmdClientService) RunCmd(width int32) ([]*pb.Replay, error) {
	req := &pb.CmdReq{Cmd: r.cmd, Timeout: r.timeout, Nodelist: r.nodelist, Width: width, Port: r.port}
	log.Debugf("Command [%s] Start ...\n", r.cmd)
	resp, err := r.client.RunCmd(r.ctx, req)
	if err != nil {
		return nil, err
	}
	log.Debugf("Command [%s] End\n", r.cmd)
	resps := make([]*pb.Replay, 0)
	if resp != nil {
		resps = append(resps, resp.Replay...)
	}
	return resps, nil
}

func (r *RunCmdClientService) Gather(replay []*pb.Replay, nodelist string, flag bool) (idleNodes, downNodes, cancelNodes []string) {
	idleNodes = make([]string, 0)
	downNodes = make([]string, 0)
	cancelNodes = make([]string, 0)
	if flag {
		// 顺序打印结果
		replayMap := make(map[string]*pb.Replay)
		for _, rep := range replay {
			nodelist := utils.ExpNodes(rep.Nodelist)
			for _, node := range nodelist {
				replayMap[node] = rep
			}
		}
		for _, node := range utils.ExpNodes(nodelist) {
			rep := replayMap[node]
			if rep.Pass {
				idleNodes = append(idleNodes, node)
				log.Infof("[%s] %s\n%s\n", log.ColorWrapper("PASS", log.Success), node, rep.Msg)
			} else {
				if rep.Msg == "canceled" {
					cancelNodes = append(cancelNodes, node)
					log.Infof("[%s] %s\n%s\n", log.ColorWrapper("CANCEL", log.Cancel), node, rep.Msg)
				} else {
					downNodes = append(downNodes, node)
					log.Infof("[%s] %s\n%s\n", log.ColorWrapper("FAILED", log.Failed), node, rep.Msg)
				}
			}
		}
	} else {
		for _, rep := range replay {
			nodelist := utils.ExpNodes(rep.Nodelist)
			if rep.Pass {
				idleNodes = append(idleNodes, nodelist...)
				log.Infof("[%s] %s\n%s\n", log.ColorWrapper("PASS", log.Success), rep.Nodelist, rep.Msg)
			} else {
				if rep.Msg == "canceled" {
					cancelNodes = append(cancelNodes, nodelist...)
					log.Infof("[%s] %s\n%s\n", log.ColorWrapper("CANCEL", log.Cancel), rep.Nodelist, rep.Msg)
				} else {
					downNodes = append(downNodes, nodelist...)
					log.Infof("[%s] %s\n%s\n", log.ColorWrapper("FAILED", log.Failed), rep.Nodelist, rep.Msg)
				}
			}
		}
	}
	return
}
