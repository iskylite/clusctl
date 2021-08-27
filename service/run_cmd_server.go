package service

import (
	"context"
	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"
	"sync"

	"google.golang.org/grpc"
)

func (p *putStreamServer) RunCmd(req *pb.CmdReq, stream pb.RpcService_RunCmdServer) error {
	// get authority
	token, _ := getAuthorityByContext(stream.Context())
	perRPCCredentials := grpc.WithPerRPCCredentials(&authority{sshKey: token})
	// init base args
	splitNodes := utils.SplitNodesByWidth(utils.ExpNodes(req.Nodelist), req.Width)
	log.Debug(splitNodes)
	repliesChannel := make(chan *pb.Reply)
	// replies handle
	var waitc sync.WaitGroup
	waitc.Add(1)
	go func() {
		defer waitc.Done()
		for reply := range repliesChannel {
			if err := stream.Send(reply); err != nil {
				log.Errorf("%s send reply into channel failed\n", reply.Nodelist)
				return
			}
			log.Debugf("%s send reply into channel ok\n", reply.Nodelist)
		}
	}()
	// global context
	localNode := utils.Hostname()
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	var wg sync.WaitGroup
	// local cmd setup
	wg.Add(1)
	go func() {
		log.Debugf("Start Command [%s]\n", req.Cmd)
		defer wg.Done()
		out, ok := utils.ExecuteShellCmdWithContext(ctx, req.Cmd)
		if !ok {
			log.Errorf("Finish Command %s, Err =\n %s", req.Cmd, string(out))
			repliesChannel <- newReply(false, out, localNode)
			return
		}
		log.Debugf("Finish Command %s, Out =\n %s", req.Cmd, string(out))
		repliesChannel <- newReply(true, string(out), localNode)
	}()

	// remote client RunCmd
	log.Debugf("Start Client Job...")
	for _, nodes := range splitNodes {
		if len(nodes) < 0 {
			continue
		}
		wg.Add(1)
		go func(nodes []string) {
			defer wg.Done()
			log.Debugf("Setup RunCmdClientService For %s\n", nodes[0])
			client, down, err := newRunCmdClientService(ctx, req.Cmd, req.Port, nodes,
				req.Width, perRPCCredentials)
			if err != nil {
				repliesChannel <- newReply(false, err.Error(), utils.Merge(nodes...))
				return
			}
			if len(down) > 0 {
				repliesChannel <- newReply(false, "connect failed", utils.Merge(down...))
			}
			defer client.CloseConn()
			client.DiscribeRepliesChannel(repliesChannel)
			client.RunCmd()
		}(nodes)
	}
	wg.Wait()
	close(repliesChannel)
	waitc.Wait()
	return nil
}
