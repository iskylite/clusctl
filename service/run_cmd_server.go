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
	replayChan := make(chan []*pb.Replay, replayBufferSize)
	log.Debugf("Result Channel Length Is %d\n", replayBufferSize)
	var wg sync.WaitGroup
	wg.Add(replayBufferSize)
	go func(wg *sync.WaitGroup, ctx context.Context) {
		log.Debugf("Start Command %s\n", req.Cmd)
		defer wg.Done()
		out, ok := utils.ExecuteShellCmdWithContext(ctx, req.Cmd)
		if !ok {
			log.Error(out)
			replayChan <- []*pb.Replay{newReplay(false, out, utils.Hostname())}
			return
		}
		log.Debugf("Command %s Finished, Out => %s", req.Cmd, string(out))
		replayChan <- []*pb.Replay{newReplay(true, string(out), utils.Hostname())}
	}(&wg, ctx)

	// batch RunCmd
	log.Debugf("Start Bench Job...")
	for _, nodes := range splitNodes {
		go func(wg *sync.WaitGroup, nodes []string) {
			defer wg.Done()
			if len(nodes) == 0 {
				log.Warning("Found Empty AllocNOdes, Skip")
				return
			}
			log.Debugf("Create RunCmdClientService For %s\n", nodes[0])
			client, err := newRunCmdClientService(ctx, req.Cmd, req.Port, nodes)
			if err != nil {
				replayChan <- []*pb.Replay{newReplay(false, err.Error(), utils.ConvertNodelist(nodes))}
				return
			}
			replays, err := client.RunCmd(req.Width)
			if err != nil {
				replayChan <- []*pb.Replay{newReplay(false, err.Error(), utils.ConvertNodelist(nodes))}
				return
			}
			replayChan <- replays
		}(&wg, nodes)
	}
	wg.Wait()
	close(replayChan)
	resps := make([]*pb.Replay, 0)
	for replays := range replayChan {
		resps = append(resps, replays...)
	}
	return &pb.PutStreamResp{Replay: resps}, nil
}
