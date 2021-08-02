package service

import (
	"fmt"
	"io"
	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"
	"os"
	"path/filepath"
	"sync"
	"time"

	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type putStreamServer struct {
	tmpDir     string
	grpcServer *grpc.Server
}

func NewPutStreamServerService(tmpDir string) *putStreamServer {
	return &putStreamServer{
		tmpDir: filepath.Join("/tmp", tmpDir),
	}
}

func (p *putStreamServer) PutStream(stream pb.RpcService_PutStreamServer) error {
	resps := make([]*pb.Replay, 0)
	// 本地默认节点生成树
	LocalNode := utils.Hostname()
	if !utils.IsDir(p.tmpDir) {
		err := os.Mkdir(p.tmpDir, 0644)
		if err != nil {
			log.Error(err.Error())
			return stream.SendAndClose(&pb.PutStreamResp{
				Replay: append(resps, newReplay(false, err.Error(), LocalNode))})
		}
	}
	// 创建临时文件
	f, err := os.CreateTemp(p.tmpDir, utils.UUID())
	if err != nil {
		log.Error(err.Error())
		return stream.SendAndClose(&pb.PutStreamResp{
			Replay: append(resps, newReplay(false, err.Error(), LocalNode)),
		})
	}
	tmpfile := f.Name()
	defer func() {
		log.Debugf("Server Send Stream [%d] Replay\n", len(resps))
		f.Close()
		if utils.Isfile(tmpfile) {
			if err = os.Remove(tmpfile); err != nil {
				log.Errorf("Server Defer Error: %s\n", err.Error())
			} else {
				log.Debugf("Server Defer Pass: remove %s\n", tmpfile)
			}
		}
		log.Debug("Server Defer End")
	}()
	cnt := 0
	var once sync.Once
	var wg sync.WaitGroup
	var fp string // 目标文件路径
	var LocalNodeList string
	nodelist := ""
	streams := make([]Wrapper, 0)
	ctx, cancel := context.WithCancel(context.Background())
	log.Debug("Server Stream Start")
	// 设置三容量的pb.PutStreamReq管道，存放当前节点和2个子节点的返回值
LOOP:
	for {
		data, err := stream.Recv()
		switch err {
		case io.EOF:
			// 流结束
			log.Debugf("Read EOF From Stream On Node [%s], cnt=[%d]\n", LocalNode, cnt)
			for _, stream := range streams {
				log.Debugf("Close DataChan On [%s]\n", stream.GetBatchNode())
				stream.CloseDataChan()
			}
			log.Debug("Wait All Stream Done ...")
			wg.Wait()
			for _, stream := range streams {
				log.Debugf("Gather replays on [%s]", stream.GetBatchNode())
				replaies, err := stream.GetResult()
				if err != nil {
					resps = append(resps, newReplay(false, err.Error(), stream.GetAllNodelist()))
				} else {
					if replaies == nil {
						resps = append(resps, newReplay(false, "no replay", stream.GetBatchNode()))
						continue
					}
					resps = append(resps, replaies.Replay...)
				}
			}
			break LOOP
		case nil:
			once.Do(func() {
				// 修改文件权限
				// 属组
				err := f.Chown(int(data.Uid), int(data.Gid))
				if err != nil {
					log.Error(err)
				}
				// 权限
				err = f.Chmod(os.FileMode(data.Filemod))
				if err != nil {
					log.Error(err)
				}
				// 修改时间
				err = os.Chtimes(tmpfile, time.Now(), time.Unix(data.Modtime, 0))
				if err != nil {
					log.Error(err)
				}
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
				fp = filepath.Join(data.Location, data.Name)
				localWriterStream, _ := newLocalWriterWrapper(ctx, fp, f, &wg)
				wg.Add(1)
				streams = append(streams, localWriterStream)
				go localWriterStream.SendFromChannel()
				// 初始化分发流客户端
				splitNodes := utils.SplitNodesByWidth(utils.ExpNodes(nodelist), data.Width)
				for _, nodes := range splitNodes {
					log.Debug(nodes)
					if len(nodes) == 0 {
						continue
					}
					addr := fmt.Sprintf("%s:%s", nodes[0], data.Port)
					stream, err := newStreamWrapper(ctx, data.Name, data.Location, data.Port, nodes, data.Width, &wg)
					if err != nil {
						log.Errorf("Server Stream Client [%s] Setup Failed\n", addr)
						resps = append(resps, newReplay(false,
							status.Code(err).String(), utils.ConvertNodelist(nodes)))
						continue
					}
					stream.SetFileInfo(data.Uid, data.Gid, data.Filemod, data.Modtime)
					wg.Add(1)
					log.Debugf("Server Stream Client [%s] Setup", addr)
					streams = append(streams, stream)
					go stream.SendFromChannel()
				}
				log.Debugf("All %d Streams Setup", len(streams))
			})
			// md5 check
			md5Str := utils.Md5sum(data.GetBody())
			if md5Str != data.Md5 {
				log.Errorf("Md5 Check Failed, cnt=[%d], md5(origin)=[%s], md5(stream)=[%s]\n",
					cnt, data.GetMd5(), md5Str)
				resps = append(resps, newReplay(false, "md5 unmatched", LocalNodeList))
				break LOOP
			}
			log.Debugf("Md5 Check Pass, cnt=[%d], md5(origin)=[%s], md5(stream)=[%s]\n",
				cnt, data.GetMd5(), md5Str)
			for _, stream := range streams {
				log.Debugf("Send Data Into Stream For Node [%s]\n", stream.GetBatchNode())
				stream.GetDataChan() <- data.GetBody()
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
				resps = append(resps, newReplay(false, err.Error(), LocalNodeList))
			}
			break LOOP
		}
	}
	cancel()
	return stream.SendAndClose(&pb.PutStreamResp{Replay: resps})
}
