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
	// run local cmd
	// cmd := utils.NewCommandWithContext(ctx, req.Cmd, int(req.GetTimeout()))
	// cmd := utils.NewCommand(req.Cmd, int(req.GetTimeout()))
	// log.Debugf("Command=[%s], Timeout=[%d]\n", req.Cmd, req.GetTimeout())
	// go cmd.Execute()
	// go func(wg *sync.WaitGroup) {
	// 	output, ok := cmd.GetResult()
	// 	log.Debugf("Local Command Output=[%s], OK=[%t]\n", output, ok)
	// 	replayChan <- []*pb.Replay{newReplay(ok, output, utils.Hostname())}
	// 	wg.Done()
	// }(&wg)
	go func(wg *sync.WaitGroup, ctx context.Context) {
		log.Debugf("Start Command=%s, Timeout=%d\n", req.Cmd, req.GetTimeout())
		defer wg.Done()
		ctx1, cancel := context.WithTimeout(ctx, utils.Timeout(int(req.Timeout)))
		defer cancel()
		out, err := utils.ExecuteShellCmdWithContext(ctx1, req.Cmd)
		if err != nil {
			log.Error(err)
			replayChan <- []*pb.Replay{newReplay(false, err.Error(), utils.Hostname())}
			return
		}
		log.Debugf("Finish Command=%s, Out=%s", req.Cmd, string(out))
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
			log.Debugf("Create RunCmdClientService For [%s]\n", nodes[0])
			client, err := newRunCmdClientService(ctx, req.Cmd, req.Port, nodes, req.Timeout)
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
