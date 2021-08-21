package service

import (
	"context"
	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"
	"sync"
)

func (p *putStreamServer) RunCmd(ctx context.Context, req *pb.CmdReq) (*pb.PutStreamResp,
	error) {
	// init
	splitNodes := utils.SplitNodesByWidth(utils.ExpNodes(req.Nodelist), req.Width)
	replayBufferSize := len(splitNodes) + 1
	resps := make([]*pb.Replay, 0)
	var wg sync.WaitGroup
	wg.Add(replayBufferSize)
	log.Debugf("WaitGroup %d\n", replayBufferSize)
	go func(wg *sync.WaitGroup, ctx context.Context) {
		log.Debugf("Start Command %s\n", req.Cmd)
		defer wg.Done()
		out, ok := utils.ExecuteShellCmdWithContext(ctx, req.Cmd)
		if !ok {
			log.Error(out)
			resps = append(resps, newReplay(false, out, utils.Hostname()))
			return
		}
		log.Debugf("Command %s Finished, Out => %s", req.Cmd, string(out))
		resps = append(resps, newReplay(true, string(out), utils.Hostname()))
	}(&wg, ctx)

	// batch RunCmd
	log.Debugf("Start Bench Job...")
	for _, nodes := range splitNodes {
		go func(wg *sync.WaitGroup, nodes []string) {
			defer wg.Done()
			if len(nodes) == 0 {
				log.Error("Found Empty AllocNOdes, Skip")
				return
			}
			log.Debugf("Create RunCmdClientService For %s\n", nodes[0])
			client, down, err := newRunCmdClientService(ctx, req.Cmd, req.Port, nodes)
			if err != nil {
				resps = append(resps, newReplay(false, err.Error(), utils.ConvertNodelist(nodes)))
				return
			}
			if len(down) > 0 {
				resps = append(resps, newReplay(false, "connect failed", utils.ConvertNodelist(down)))
			}
			defer client.CloseConn()
			replays, err := client.RunCmd(req.Width)
			if err != nil {
				resps = append(resps, newReplay(false, err.Error(), utils.ConvertNodelist(nodes)))
				return
			}
			resps = append(resps, replays...)
		}(&wg, nodes)
	}
	wg.Wait()
	return &pb.PutStreamResp{Replay: resps}, nil
}
