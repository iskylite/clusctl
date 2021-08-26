package service

import (
	"context"
	"fmt"
	"myclush/global"
	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/iskylite/nodeset"
	"google.golang.org/grpc"
)

type PingClientService struct {
	nodelist  string
	replyChan chan *pb.Reply
	workers   int
	port      string
	timeout   int
}

func NewPingClientService(nodelist, port string, workers int) *PingClientService {
	return &PingClientService{
		nodelist:  nodelist,
		replyChan: make(chan *pb.Reply, runtime.NumCPU()*2),
		workers:   workers,
		port:      port,
	}
}

func (p *PingClientService) SetTimeout(timeout int) {
	p.timeout = timeout
}

func (p *PingClientService) Ping(ctx context.Context, node string) {
	addr := fmt.Sprintf("%s:%s", node, p.port)
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(p.timeout))
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithBlock(), grpc.WithInsecure(), global.Authority)
	defer func() {
		if conn == nil {
			return
		}
		err := conn.Close()
		if err != nil {
			log.Error(utils.GrpcErrorMsg(err))
		}
	}()
	if err != nil {
		p.replyChan <- newReply(false, utils.GrpcErrorMsg(err), node)
		log.Errorf("PingError %s: %s\n", node, utils.GrpcErrorMsg(err))
	} else {
		client := pb.NewRpcServiceClient(conn)
		reply, err := client.Ping(ctx, &pb.CommonReq{Version: global.Version})
		if err != nil {
			p.replyChan <- newReply(false, utils.GrpcErrorMsg(err), node)
			log.Errorf("PingError %s: %s\n", node, utils.GrpcErrorMsg(err))
		} else {
			if reply.GetOk() {
				p.replyChan <- newReply(true, global.Success, node)
			} else {
				p.replyChan <- newReply(false, "Version Unmatched", node)
				log.Errorf("PingError %s: Version Unmatched\n", node)
			}
		}
	}
}

func (p *PingClientService) PingFromChan(ctx context.Context, nodeChan chan string, wg *sync.WaitGroup) {
	for node := range nodeChan {
		p.Ping(ctx, node)
	}
	wg.Done()
}

func (p *PingClientService) Run(ctx context.Context) error {
	defer close(p.replyChan)
	numLimit := runtime.NumCPU()
	nodeChan := make(chan string, numLimit)
	var wg sync.WaitGroup
	wg.Add(p.workers)
	for i := 0; i < p.workers; i++ {
		log.Debugf("start Ping Worker [%d]\n", i)
		go p.PingFromChan(ctx, nodeChan, &wg)
	}
	iter, err := nodeset.Yield(p.nodelist)
	if err != nil {
		close(nodeChan)
		return err
	}
	for iter.Next() {
		nodeChan <- iter.Value()
	}
	close(nodeChan)
	wg.Wait()
	return nil
}

func (p *PingClientService) Gather() {
	idleNodes := make([]string, 0)
	downNodes := make([]string, 0)
	for rep := range p.replyChan {
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
