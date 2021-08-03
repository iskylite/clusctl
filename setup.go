package main

import (
	"context"
	log "myclush/logger"
	"myclush/pb"
	"myclush/service"
	"myclush/utils"
	"strconv"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func setLogLevel(debug bool) {
	if debug {
		log.SetLevel(log.DEBUG)
		log.Debug("Logger Setup In DEBUG Mode")
	} else {
		log.SetSilent()
	}
}

func putStreamClientServiceSetup(ctx context.Context, cancel func(), localFile, destDir, nodes, buffer string, port, width int) {
	defer cancel()
	bufferSize, err := utils.ConvertSize(buffer)
	if err != nil {
		log.Error(err)
		return
	}
	clientService, err := service.NewPutStreamClientService(localFile, destDir, nodes, strconv.Itoa(port), int32(width))
	if err != nil {
		log.Errorf("PutStreamClientService Failed, err=[%s]\n", err.Error())
		return
	}
	err = clientService.GenStreamWithContext(ctx)
	if err != nil {
		log.Errorf("PutStreamClientService Failed, err=[%s]\n",
			status.Code(err).String())
		return
	}
	err = clientService.RunServe(ctx, bufferSize)
	if err != nil {
		log.Errorf("PutStreamClientService Failed, err=[%s], T=[%#s]\n", err.Error(), status.Code(err).String())
		// 取消或者发送失败需要汇总错误信息
		// return
	}
	replays, err := clientService.CloseAndRecv()
	log.Debug("PutStreamClientService Recv All Replay...")
	if err != nil {
		if status.Code(err) == codes.Canceled {
			log.Errorf("PutStreamClientService Canceled\n")
		} else {
			log.Errorf("PutStreamClientService Failed, err=[%s]\n", err.Error())
		}
		return
	}

	idleNodes, downNodes, cancelNodes := clientService.Gather(replays.Replay)

	log.Infof("%s: %s, %s: %d\n", log.ColorWrapper("PASS", log.Success), utils.ConvertNodelist(idleNodes), log.ColorWrapper("SUM", log.Success), len(idleNodes))
	if len(downNodes) > 0 {
		log.Infof("%s: %s, %s: %d\n", log.ColorWrapper("FAILED", log.Failed), utils.ConvertNodelist(downNodes), log.ColorWrapper("SUM", log.Failed), len(downNodes))
	}
	if len(cancelNodes) > 0 {
		log.Infof("%s: %s, %s: %d\n", log.ColorWrapper("CANCEL", log.Cancel), utils.ConvertNodelist(cancelNodes), log.ColorWrapper("SUM", log.Cancel), len(cancelNodes))
	}
	log.Debug("PutStreamClientService Stop")
}

func putStreamServerServiceSetup(ctx context.Context, cancel func(), tmpDir string, port int) {
	serverService := service.NewPutStreamServerService(tmpDir)
	go func() {
		defer cancel()
		err := serverService.RunServer(strconv.Itoa(port))
		if err != nil {
			log.Errorf("PutStreamServerService Failed, err=[%s]\n", err.Error())
			return
		}
	}()
	<-ctx.Done()
	serverService.Stop()
	log.Info("PutStreamServerService Stop")
}

func RunCmdClientServiceSetup(ctx context.Context, cancel context.CancelFunc, cmd, nodes string, width, port int, list bool) {
	defer cancel()
	client, err := service.NewRunCmdClientService(ctx, cmd, nodes, strconv.Itoa(port))
	if err != nil {
		log.Error(err)
		return
	}
	replays, err := client.RunCmd(int32(width))
	if err != nil {
		log.Error(err)
		return
	}
	idleNodes, downNodes, cancelNodes := client.Gather(replays, nodes, list)

	log.Infof("%s: %s, %s: %d\n", log.ColorWrapper("PASS", log.Success), utils.ConvertNodelist(idleNodes), log.ColorWrapper("SUM", log.Success), len(idleNodes))
	if len(downNodes) > 0 {
		log.Infof("%s: %s, %s: %d\n", log.ColorWrapper("FAILED", log.Failed), utils.ConvertNodelist(downNodes), log.ColorWrapper("SUM", log.Failed), len(downNodes))
	}
	if len(cancelNodes) > 0 {
		log.Infof("%s: %s, %s: %d\n", log.ColorWrapper("CANCEL", log.Cancel), utils.ConvertNodelist(cancelNodes), log.ColorWrapper("SUM", log.Cancel), len(cancelNodes))
	}
}

func PingClientServiceSetup(ctx context.Context, nodes string, port, workers, timeout int) {
	pingClientService := service.NewPingClientService(nodes, strconv.Itoa(port), workers)
	pingClientService.SetTimeout(timeout)
	go pingClientService.Run(ctx)
	idleNodes, downNodes := pingClientService.Gather()

	log.Infof("%s: %s, %s: %d\n", log.ColorWrapper("PASS", log.Success), utils.ConvertNodelist(idleNodes), log.ColorWrapper("SUM", log.Success), len(idleNodes))
	if len(downNodes) > 0 {
		log.Infof("%s: %s, %s: %d\n", log.ColorWrapper("FAILED", log.Failed), utils.ConvertNodelist(downNodes), log.ColorWrapper("SUM", log.Failed), len(downNodes))
	}
}

func OoBwServiceSetup(ctx context.Context, nodes, nnodes, oobwbuffer string, after int64, oobwloop bool, port, length, count int) {
	cnodes := utils.ExpNodes(nodes)
	cnode_len := len(cnodes)
	snodes := utils.ExpNodes(nnodes)
	snode_len := len(snodes)
	if cnode_len != snode_len {
		log.Error("client nodes num not equal server nodes num")
		return
	}
	results := make(chan *pb.Replay, cnode_len)
	var wg sync.WaitGroup
	timer := utils.NewTimerAfterSeconds(after)
	wg.Add(cnode_len)
	for index, node := range cnodes {
		go service.RunOoBwClientService(ctx, snodes[index], node, oobwbuffer, strconv.Itoa(port), timer, results, &wg, oobwloop, length, count)
	}
	wg.Wait()
	close(results)
	if oobwloop {
		var passAvgSum float32
		var passCnt int
		transLength := 8 << length
		log.Infof("[State] Nodes\tOffset\tLoop\tAvgBw(MB/s)\n")
		for replay := range results {
			if replay.Pass {
				avg, err := service.ParseOoBwResult(replay.Msg)
				if err != nil {
					log.Infof("[%s] %s\t%d\t%d\t%s\n", log.ColorWrapper("FAILED", log.Failed), replay.Nodelist, length, count, avg)
					continue
				}
				log.Infof("[%s] %s\t%d\t%d\t%.2f\n", log.ColorWrapper("PASS", log.Success), replay.Nodelist, length, count, avg)
				passAvgSum += avg
				passCnt++
				continue
			}
			log.Infof("[%s] %s\t%d\t%d\t%s\n", log.ColorWrapper("FAILED", log.Failed), replay.Nodelist, length, count, replay.Msg)
		}

		log.Infof("\n[%s] %d\t%d\t%d\t%.2f\n", log.ColorWrapper("PASS", log.Success), passCnt, transLength, count, passAvgSum)
	} else {
		for replay := range results {
			if replay.Pass {
				log.Infof("[%s] %s\n%s\n", log.ColorWrapper("PASS", log.Success), replay.Nodelist, replay.Msg)
				continue
			}
			log.Infof("[%s] %s\n%s\n", log.ColorWrapper("FAILED", log.Failed), replay.Nodelist, replay.Msg)
		}
	}

}
