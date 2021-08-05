package service

import (
	"context"
	"errors"
	"fmt"
	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"

	"google.golang.org/grpc"
)

func (p *putStreamServer) RunOoBwClient(ctx context.Context, req *pb.OoBwClientReq) (*pb.Replay, error) {
	// 获取本地ncid
	localNode := fmt.Sprintf("%s -> %s", utils.Hostname(), req.Server)
	empty := newReplay(false, "nil", localNode)
	localNcid, err := getLocalNcidWithContext(ctx)
	if err != nil {
		log.Error(err)
		return empty, err
	}
	log.Infof("Local Ncid is [%s]\n", localNcid)
	// 建立服务端连接
	addr := fmt.Sprintf("%s:%s", req.Server, req.Port)
	conn, err := grpc.DialContext(ctx, addr, grpc.WithBlock(), grpc.WithInsecure())
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Error(err)
		}
	}()
	if err != nil {
		return newReplay(false, err.Error(), localNode), err
	}
	client := pb.NewRpcServiceClient(conn)
	// 获取服务端ncid
	serverNcid := ""
	replay, err := client.GetNcid(ctx, &pb.GG{HH: req.Server})
	if err != nil {
		log.Error(err)
		return empty, err
	}
	serverNcid = replay.Msg
	log.Infof("Server Ncid is [%s]\n", serverNcid)
	// 开启服务端
	ctxs, cancel := context.WithCancel(ctx)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(1)
	go func(ctx context.Context, wg *sync.WaitGroup, length, count int32) {
		defer wg.Done()
		replay, err := client.RunOoBwServer(ctx, &pb.OoBwServerReq{
			Ncid:   localNcid,
			Buffer: req.Buffer,
			Length: length,
			Count:  count,
		})
		if err != nil {
			log.Error(err)
		} else {
			log.Debugf("node %s server pass %t\n", replay.Nodelist, replay.Pass)
		}
	}(ctxs, &wg, req.Length, req.Count)
	// 启动定时器
	log.Debugf("Timer To %s\n", utils.FormatTime(req.Timer))
	// dateNanoUnix := utils.NewTimerAfterSeconds(req.Timer)
	// file test
	var cmdFile, args string
	if req.Length != 0 && req.Count != 0 {
		cmdFile = "/usr/local/glex/examples/oo_bw_s_loop"
		args = fmt.Sprintf("8 %s 8 %s %d %d", serverNcid, req.Buffer, req.Length, req.Count)
	} else {
		cmdFile = "/usr/local/glex/examples/oo_bw_s"
		args = fmt.Sprintf("8 %s 8 %s", serverNcid, req.Buffer)
	}
	if !utils.Isfile(cmdFile) {
		log.Errorf("%s not exist", cmdFile)
		return empty, fmt.Errorf("%s not exist", cmdFile)
	}
	execname := filepath.Base(cmdFile)
	// 启动定时器
	timer := utils.GenTikerWithTimer(req.Timer)
	<-timer
	log.Debugf("%s start on %s\n", execname, utils.LocalTime())
	// 启动客户端RunOoBwClient
	command := fmt.Sprintf("%s %s", cmdFile, args)
	out, ok := utils.ExecuteShellCmdWithContext(ctx, command)
	defer log.Debugf("%s finish on %s\n", execname, utils.LocalTime())
	if !ok {
		log.Error(out)
		return empty, errors.New(out)
	}
	wg.Wait()
	return newReplay(true, string(out), localNode), nil
}

func RunOoBwClientService(ctx context.Context, server, node, buffer, port string, timer int64, results chan *pb.Replay, wg *sync.WaitGroup, oobwloop bool, length, count int) {
	defer wg.Done()
	addr := fmt.Sprintf("%s:%s", node, port)
	if !oobwloop {
		length = 0
		count = 0
	}
	log.Debugf("Length=%d, Loop=%d\n", length, count)
	conn, err := grpc.DialContext(ctx, addr, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		results <- newReplay(false, utils.GrpcErrorMsg(err), node)
		return
	}
	client := pb.NewRpcServiceClient(conn)
	log.Debugf("oo_bw test start on %s -> %s\n", node, server)
	replay, err := client.RunOoBwClient(ctx, &pb.OoBwClientReq{
		Server: server,
		Buffer: buffer,
		Timer:  timer,
		Port:   port,
		Length: int32(length),
		Count:  int32(count),
	})
	if err != nil {
		results <- newReplay(false, utils.GrpcErrorMsg(err), node)
		return
	}
	results <- replay
}

func ParseOoBwResult(result string) (float32, error) {
	var avg float32
	re := regexp.MustCompile(`.*\d+\s+\d+\s+[0-9\.]+\s+([0-9\.]+)\s+OK.*`)
	matches := re.FindAllStringSubmatch(result, -1)
	if len(matches) == 0 {
		return avg, fmt.Errorf("no matches")
	}
	cnts := len(matches)
	for _, match := range matches {
		write, err := strconv.ParseFloat(match[1], 32)
		if err != nil {
			return avg, err
		}
		avg += float32(write)
	}
	return avg / float32(cnts), nil
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
		go RunOoBwClientService(ctx, snodes[index], node, oobwbuffer, strconv.Itoa(port), timer, results, &wg, oobwloop, length, count)
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
				avg, err := ParseOoBwResult(replay.Msg)
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
