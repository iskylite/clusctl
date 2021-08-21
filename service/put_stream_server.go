package service

import (
	"fmt"
	"io"
	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"
	"runtime"
	"sync"

	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (p *putStreamServer) PutStream(stream pb.RpcService_PutStreamServer) error {
	cnt := 0
	var once sync.Once
	var wg sync.WaitGroup
	var LocalNodeList string
	nodelist := ""
	streams := make([]Wrapper, 0)
	defer func() {
		for _, stream := range streams {
			stream.CloseConn()
		}
		log.Info("stop putStream server")
	}()
	ctx, cancel := context.WithCancel(context.Background())
	log.Debug("PutStream Server Start ... ")
	// 响应值通道
	replaiesChannel := make(chan *pb.Replay, runtime.NumCPU())
	var waitc sync.WaitGroup
	waitc.Add(1)
	go func() {
		defer waitc.Done()
		for replay := range replaiesChannel {
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
			close(replaiesChannel)
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
					replaiesChannel <- newReplay(false, err.Error(), LocalNode)
				} else {
					wg.Add(1)
					streams = append(streams, localWriterStream)
					go localWriterStream.SendFromChannel()
					localWriterStream.DiscribeReplaiesChannel(replaiesChannel)
				}
				// 初始化分发流客户端
				splitNodes := utils.SplitNodesByWidth(utils.ExpNodes(nodelist), data.Width)
				for _, nodes := range splitNodes {
					log.Debug(nodes)
					if len(nodes) == 0 {
						continue
					}
					addr := fmt.Sprintf("%s:%s", nodes[0], data.Port)
					// 只要有一个连接成功就不会返回错误
					stream, down, err := newStreamWrapper(ctx, data.Name, data.Location, data.Port, nodes, data.Width, &wg)
					if err != nil {
						log.Errorf("Server Stream Client [%s] Setup Failed\n", addr)
						replaiesChannel <- newReplay(false,
							utils.GrpcErrorMsg(err), utils.ConvertNodelist(nodes))
						continue
					}
					if len(down) > 0 {
						replaiesChannel <- newReplay(false, "connect timeout or failed", utils.ConvertNodelist(down))
					}
					stream.SetFileInfo(data.Uid, data.Gid, data.Filemod, data.Modtime)
					wg.Add(1)
					log.Debugf("Server Stream Client [%s] Setup", addr)
					streams = append(streams, stream)
					stream.DiscribeReplaiesChannel(replaiesChannel)
					go stream.SendFromChannel()
				}
				log.Debugf("All %d Streams Setup", len(streams))
			})
			// md5 check
			md5Str := utils.Md5sum(data.GetBody())
			if md5Str != data.Md5 {
				log.Errorf("Md5 Check Failed, cnt=[%d], md5(origin)=[%s], md5(stream)=[%s]\n",
					cnt, data.GetMd5(), md5Str)
				replaiesChannel <- newReplay(false, "md5 unmatched", LocalNodeList)
				break LOOP
			}
			for _, stream := range streams {
				// if stream.IsLocal() {
				// 	log.Debug("Send Data Into Stream For Local")
				// } else {
				// 	log.Debugf("Send Data Into Stream For [%s]\n", stream.GetBatchNode())
				// }
				stream.RecvData(data.GetBody())
			}
			cnt++
		default:
			// Error Handler
			if LocalNodeList == "" {
				LocalNodeList = fmt.Sprintf("%s,%s", LocalNode, nodelist)
			}
			if status.Code(err) == codes.Canceled {
				log.Error("Stream Recv Canceled Signal")
				cancel()
				return nil
			} else {
				log.Errorf("stream recv error: [%s]\n", err.Error())
				replaiesChannel <- newReplay(false, utils.GrpcErrorMsg(err), LocalNodeList)
			}
			break LOOP
		}
	}
	// cancel()
	return nil
}
