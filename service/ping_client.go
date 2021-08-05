package service

import (
	"context"
	"fmt"
	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"
	"runtime"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
)

type PingClientService struct {
	nodelist   string
	replayChan chan *pb.Replay
	workers    int
	port       string
	timeout    int
}

func NewPingClientService(nodelist, port string, workers int) *PingClientService {
	return &PingClientService{
		nodelist:   nodelist,
		replayChan: make(chan *pb.Replay, runtime.NumCPU()*4),
		workers:    workers,
		port:       port,
	}
}

func (p *PingClientService) SetTimeout(timeout int) {
	p.timeout = timeout
}

func (p *PingClientService) Ping(ctx context.Context, node string) {
	addr := fmt.Sprintf("%s:%s", node, p.port)
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(p.timeout))
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithBlock(), grpc.WithInsecure())
	defer func() {
		if conn == nil {
			return
		}
		err := conn.Close()
		if err != nil {
			log.Error(err)
		}
	}()
	if err != nil {
		p.replayChan <- newReplay(false, utils.GrpcErrorMsg(err), node)
		log.Errorf("Node [%s] Error: %s\n", node, err)
	} else {
		client := pb.NewRpcServiceClient(conn)
		replay, err := client.Ping(ctx, &pb.GG{HH: utils.Hostname()})
		if err != nil {
			p.replayChan <- newReplay(false, utils.GrpcErrorMsg(err), node)
			log.Error(err)
		} else {
			p.replayChan <- newReplay(true, "", replay.GetHH())
		}
	}
}

func (p *PingClientService) PingFromChan(ctx context.Context, nodeChan chan string, wg *sync.WaitGroup) {
	for node := range nodeChan {
		p.Ping(ctx, node)
	}
	wg.Done()
}

func (p *PingClientService) Run(ctx context.Context) {
	defer close(p.replayChan)
	numLimit := runtime.NumCPU()
	nodeChan := make(chan string, numLimit)
	go utils.AddNode(p.nodelist, nodeChan)
	var wg sync.WaitGroup
	wg.Add(p.workers)
	for i := 0; i < p.workers; i++ {
		log.Debugf("start Ping Worker [%d]\n", i)
		go p.PingFromChan(ctx, nodeChan, &wg)
	}
	wg.Wait()
}

func (p *PingClientService) Gather() {
	idleNodes := make([]string, 0)
	downNodes := make([]string, 0)
	for rep := range p.replayChan {
		if rep.Pass {
			idleNodes = append(idleNodes, rep.Nodelist)
		} else {
			downNodes = append(downNodes, rep.Nodelist)
		}
	}
	if len(idleNodes) > 0 {
		log.ColorWrapperInfo(log.Success, idleNodes, "")
	}
	if len(downNodes) > 0 {
		log.ColorWrapperInfo(log.Failed, downNodes, "")
	}
}

func PingClientServiceSetup(ctx context.Context, nodes string, port, workers, timeout int) {
	pingClientService := NewPingClientService(nodes, strconv.Itoa(port), workers)
	pingClientService.SetTimeout(timeout)
	go pingClientService.Run(ctx)
	pingClientService.Gather()
}
