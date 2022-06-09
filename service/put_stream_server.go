package service

import (
	"fmt"
	"io"
	"myclush/global"
	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"
	"runtime"
	"sync"

	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (p *putStreamServer) PutStream(stream pb.RpcService_PutStreamServer) error {
	if global.MunalGC {
		defer log.Info("runtime.GC() end!")
		defer runtime.GC()
	}
	// get authority
	token, _ := getAuthorityByContext(stream.Context())
	perRPCCredentials := grpc.WithPerRPCCredentials(&authority{sshKey: token})
	cnt := 0
	var once sync.Once
	var wg sync.WaitGroup
	var LocalNodeList string
	var LocalNode string
	nodelist := ""
	streams := make([]Wrapper, 0)
	defer func() {
		for _, stream := range streams {
			stream.CloseConn()
		}
		log.Info("PutStream Server Finished !!!")
	}()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log.Debug("PutStream Server Start ... ")
	// 响应值通道
	repliesChannel := make(chan *pb.Reply, runtime.NumCPU())
	var waitc sync.WaitGroup
	waitc.Add(1)
	go func() {
		defer waitc.Done()
		for replay := range repliesChannel {
			if err := stream.Send(replay); err != nil {
				log.Errorf("node=%s, send replay=%s, %v\n", replay.Nodelist, replay.Msg, err)
				break
			}
			log.Debugf("node=%s, send replay=%s\n", replay.Nodelist, replay.Msg)
		}
	}()
LOOP:
	for {
		data, err := stream.Recv()
		// 替代本地主机名
		LocalNode = data.GetNode()
		switch err {
		case io.EOF:
			// 流结束
			log.Debugf("cnt=[%d], recv io.EOF\n", cnt)
			for _, stream := range streams {
				stream.CloseDataChan()
				log.Debugf("node=%s, Close DataChan\n", stream.GetBatchNode())
			}
			log.Debug("Wait All Stream Done ...")
			wg.Wait()
			// 处理响应
			close(repliesChannel)
			waitc.Wait()
			break LOOP
		case nil:
			once.Do(func() {
				// 根据流初始化默认配置
				nodelist = data.Nodelist
				if nodelist == "" {
					LocalNodeList = LocalNode
				} else {
					LocalNodeList = fmt.Sprintf("%s,%s", LocalNode, nodelist)
				}
				if nodelist != "" {
					log.Debugf("Next Allocte NodeList [%s] By Width [%d]\n", nodelist, data.Width)
				}
				// 本地数据写入流客户端
				localWriterStream, err := newLocalWriterWrapper(ctx, data, p.tmpDir, &wg)
				if err != nil {
					repliesChannel <- newReply(false, err.Error(), LocalNode)
				} else {
					wg.Add(1)
					streams = append(streams, localWriterStream)
					go localWriterStream.SendFromChannel()
					localWriterStream.DiscribeRepliesChannel(repliesChannel)
				}
				// 初始化分发流客户端
				splitNodes := utils.SplitNodesByWidth(utils.ExpNodes(nodelist), data.Width)
				for _, nodes := range splitNodes {
					if len(nodes) == 0 {
						continue
					}
					// 只要有一个连接成功就不会返回错误
					stream, down, node, err := newStreamWrapper(ctx, data.Name, data.Location, data.Port, nodes, data.Width, &wg, perRPCCredentials)
					if err != nil {
						log.Errorf("Server Stream Client [%s] Setup Failed\n", utils.Merge(nodes...))
						repliesChannel <- newReply(false,
							utils.GrpcErrorMsg(err), utils.Merge(nodes...))
						continue
					}
					if len(down) > 0 {
						repliesChannel <- newReply(false, "rpc timeout or failed", utils.Merge(down...))
					}
					stream.SetFileInfo(data.Uid, data.Gid, data.Filemod, data.Modtime)
					wg.Add(1)
					log.Infof("Server Stream Client [%s] Setup Success", node)
					streams = append(streams, stream)
					stream.DiscribeRepliesChannel(repliesChannel)
					go stream.SendFromChannel()
				}
				log.Debugf("All %d Streams Setup", len(streams))
			})
			// md5 check
			md5Str := utils.Md5sum(data.GetBody())
			if md5Str != data.Md5 {
				log.Errorf("Md5 Check Failed, cnt=[%d], md5(origin)=[%s], md5(stream)=[%s]\n",
					cnt, data.GetMd5(), md5Str)
				repliesChannel <- newReply(false, "md5 unmatched", LocalNodeList)
				break LOOP
			}
			for _, stream := range streams {
				// if stream.IsLocal() {
				// 	log.Debugf("Send Data Into Stream For Local")
				// } else {
				// 	log.Debugf("Send Data Into Stream For [%s]", stream.GetBatchNode())
				// }
				stream.RecvData(data.GetBody())
			}
			cnt++
		default:
			// Error Handler
			if LocalNodeList == "" {
				LocalNodeList = fmt.Sprintf("%s,%s", LocalNode, nodelist)
			}
			defer func() {
				close(repliesChannel)
				log.Info("close repliesChannel")
			}()
			if status.Code(err) == codes.Canceled {
				for _, stream := range streams {
					stream.SetBad()
				}
				log.Info("gRPC receive canceled signal, cancel all stream")
				cancel()
				for _, stream := range streams {
					stream.CloseDataChan()
					stream.CleanDataChan()
				}
				return nil
			}
			log.Errorf("stream recv error: [%s]\n", err.Error())
			repliesChannel <- newReply(false, utils.GrpcErrorMsg(err), LocalNodeList)

			for _, stream := range streams {
				stream.SetBad()
				stream.CloseDataChan()
				stream.CleanDataChan()
			}

			break LOOP
		}
	}
	// cancel()
	return nil
}
