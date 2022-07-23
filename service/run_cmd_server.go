package service

import (
	"context"
	"myclush/global"
	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"
	"sync"

	"google.golang.org/grpc"
)

func (p *putStreamServer) RunCmd(req *pb.CmdReq, stream pb.RpcService_RunCmdServer) error {
	// global context
	localNode := req.GetNode()
	// get authority
	token, _ := getAuthorityByContext(stream.Context())
	perRPCCredentials := grpc.WithPerRPCCredentials(&authority{sshKey: token})
	// init base args
	log.Debug(req.Nodelist)
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
				log.Errorf("%s send reply from channel failed\n", reply.Nodelist)
				return
			}
			log.Debugf("%s send reply from channel ok\n", reply.Nodelist)
		}
	}()
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	var wg sync.WaitGroup
	// local cmd setup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if req.Daemon {
			log.Debugf("Start Daemon Command [%s]\n", req.Cmd)
			if cmd, err := utils.ExecuteShellCmdDaemon(req.Cmd); err != nil {
				log.Errorf("Finish Command %s, Err =\n\t[%v]", req.Cmd, err)
				repliesChannel <- newReply(false, err.Error(), localNode)
			} else {
				log.Debugf("Finish Command %s\n", req.Cmd)
				// wait for daemon process to exit, fix bash defunct process
				go cmd.Wait()
				repliesChannel <- newReply(true, global.SUCCESS, localNode)
			}
		} else {
			log.Debugf("Start Command [%s]\n", req.Cmd)
			out, ok := utils.ExecuteShellCmdWithContext(ctx, req.Cmd)
			if !ok {
				log.Errorf("Finish Command %s, Err =\n\t[%s]", req.Cmd, string(out))
				repliesChannel <- newReply(false, out, localNode)
				return
			}
			log.Debugf("Finish Command %s, Out =\n\t[%s]", req.Cmd, string(out))
			repliesChannel <- newReply(true, string(out), localNode)
		}
	}()

	// remote client RunCmd
	log.Debugf("Start Client Job...")
	for _, nodes := range splitNodes {
		if len(nodes) == 0 {
			continue
		}
		wg.Add(1)
		go func(nodes []string) {
			defer wg.Done()
			client, down, err := newRunCmdClientService(ctx, req.Cmd, req.Port, nodes,
				req.Width, perRPCCredentials, req.Daemon)
			if err != nil {
				log.Errorf("Setup RunCmdClientService For %s failed\n", nodes[0])
				repliesChannel <- newReply(false, utils.GrpcErrorMsg(err), utils.Merge(nodes...))
				return
			}
			log.Infof("Setup RunCmdClientService For %s Success\n", nodes[0])
			if len(down) > 0 {
				repliesChannel <- newReply(false, "rpc timeout or failed", utils.Merge(down...))
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
