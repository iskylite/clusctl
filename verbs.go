package main

import (
	"context"
	"flag"
	"fmt"
	log "myclush/logger"
	"myclush/pb"
	"myclush/service"
	"myclush/utils"
	"runtime"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	VERSION string = "v1.4.0"
	APP     string = "myclush"
)

const (
	CLIENT = 1 << iota
	SERVER
	EXECUTE
	PING
	OO
)

var (
	cc uint
	ss uint
	ee uint
	pp uint
	oo uint
)

var (
	execute    string
	client     string
	dest       string
	nodes      string
	nnodes     string
	port       string
	buffer     int
	debug      bool
	server     bool
	width      int
	timeout    int
	workers    int
	ping       bool
	list       bool
	oobw       bool
	oobwloop   bool
	after      int64
	oobwbuffer string
	length     int
	loop       int
)

func init() {
	flag.BoolVar(&server, "s", false, "start myclush server service")
	flag.StringVar(&client, "c", "", "start myclush client and copy file to remote server")
	flag.IntVar(&buffer, "b", 1024*512, "buffersize bytes")
	flag.BoolVar(&debug, "D", false, "debug log")
	flag.StringVar(&nodes, "n", "", "nodes string")
	flag.StringVar(&nnodes, "N", "", "dest nodes string")
	flag.StringVar(&dest, "d", "/tmp", "destPath")
	flag.StringVar(&port, "p", "1995", "grpc server port")
	flag.IntVar(&width, "w", 2, "B tree width")
	flag.StringVar(&execute, "e", "", "command string")
	flag.IntVar(&timeout, "t", 3, "command execute timeout")
	flag.BoolVar(&ping, "P", false, "start ping service")
	flag.BoolVar(&oobw, "o", false, "start oo_bw service")
	flag.BoolVar(&oobwloop, "O", false, "start oo_bw loop service")
	flag.IntVar(&workers, "W", runtime.NumCPU(), "ping  workers max number")
	flag.BoolVar(&list, "l", false, "sort cmd output by node list")
	flag.Int64Var(&after, "a", 3, "oobw run after seconds")
	flag.StringVar(&oobwbuffer, "B", "0x1000000", "oobw block size")
	flag.IntVar(&length, "ln", 15, "oo_bw_loop offset length")
	flag.IntVar(&loop, "lp", 100, "oo_bw_loop loop count")
	flag.Usage = func() {
		fmt.Printf("\nName: \t%s \nVersion: %s\n\nOptions:\n", APP, VERSION)
		flag.PrintDefaults()
	}
	flag.Parse()
	if client != "" {
		cc = CLIENT
		if !debug {
			log.SetSilent()
		}
	}
	if server {
		ss = SERVER
	}
	if execute != "" {
		ee = EXECUTE
		if !debug {
			log.SetSilent()
		}
	}
	if ping {
		pp = PING
		if !debug {
			log.SetSilent()
		}
	}
	if oobw || oobwloop {
		oo = OO
		if !debug {
			log.SetSilent()
		}
		if !oobwloop {
			length = 0
			loop = 0
		}
	}
	log.SetColor()
}

func putStreamClientServiceSetup(ctx context.Context, cancel func()) {
	log.Debugf("PutStreamClientService [%s] Start...\n", VERSION)
	defer cancel()
	clientService, err := service.NewPutStreamClientService(client, dest, nodes, port, int32(width))
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
	err = clientService.RunServe(ctx, buffer)
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

func putStreamServerServiceSetup(ctx context.Context, cancel func()) {
	serverService := service.NewPutStreamServerService(APP)
	go func() {
		defer cancel()
		log.Infof("PutStreamServerService [%s] Start ...\n", VERSION)
		err := serverService.RunServer(port)
		if err != nil {
			log.Errorf("PutStreamServerService Failed, err=[%s]\n", err.Error())
			return
		}
	}()
	<-ctx.Done()
	serverService.Stop()
	log.Info("PutStreamServerService Stop")
}

func RunCmdClientServiceSetup(ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	log.Debugf("RunCmdClientService [%s] start ...\n", VERSION)
	client, err := service.NewRunCmdClientService(ctx, execute, nodes, port, int32(timeout))
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

func PingClientServiceSetup(ctx context.Context) {
	pingClientService := service.NewPingClientService(nodes, port, workers)
	pingClientService.SetTimeout(timeout)
	go pingClientService.Run(ctx)
	idleNodes, downNodes := pingClientService.Gather()

	log.Infof("%s: %s, %s: %d\n", log.ColorWrapper("PASS", log.Success), utils.ConvertNodelist(idleNodes), log.ColorWrapper("SUM", log.Success), len(idleNodes))
	if len(downNodes) > 0 {
		log.Infof("%s: %s, %s: %d\n", log.ColorWrapper("FAILED", log.Failed), utils.ConvertNodelist(downNodes), log.ColorWrapper("SUM", log.Failed), len(downNodes))
	}
}

func OoBwServiceSetup(ctx context.Context) {
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
		go service.RunOoBwClientService(ctx, snodes[index], node, oobwbuffer, port, int32(timeout), timer, results, &wg, oobwloop, length, loop)
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
					log.Infof("[%s] %s\t%d\t%d\t%s\n", log.ColorWrapper("FAILED", log.Failed), replay.Nodelist, length, loop, avg)
					continue
				}
				log.Infof("[%s] %s\t%d\t%d\t%.2f\n", log.ColorWrapper("PASS", log.Success), replay.Nodelist, length, loop, avg)
				passAvgSum += avg
				passCnt++
				continue
			}
			log.Infof("[%s] %s\t%d\t%d\t%s\n", log.ColorWrapper("FAILED", log.Failed), replay.Nodelist, length, loop, replay.Msg)
		}

		log.Infof("\n[%s] %d\t%d\t%d\t%.2f\n", log.ColorWrapper("PASS", log.Success), passCnt, transLength, loop, passAvgSum)
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
