package service

import (
	"context"
	"fmt"
	"io"
	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type PutStreamClientService struct {
	filename string
	srcPath  string
	destPath string
	port     string
	nodelist string
	node     string
	width    int32
	uid      uint32
	gid      uint32
	filemod  uint32
	modtime  int64
	stream   pb.RpcService_PutStreamClient
}

func NewPutStreamClientService(fp, dp, nodelist, port string, width int32) (*PutStreamClientService, error) {
	// 判断文件是否存在
	if !utils.Isfile(fp) {
		return nil, fmt.Errorf("[%s] not found", fp)
	}
	srcPath, err := filepath.Abs(fp)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	log.Debugf("Ready For Remote Copy [%s]\n", srcPath)
	nodes := utils.ExpNodes(nodelist)
	if len(nodes) > 1 {
		nodelist = utils.ConvertNodelist(nodes[1:])
	} else if len(nodes) == 0 {
		return nil, fmt.Errorf("ExpNodes Error, nodes is empty slice")
	} else {
		nodelist = ""
	}
	node := nodes[0]
	log.Debugf("Batch Node Is [%s], All Nodes Is [%d], Width Is [%d]\n", node, len(nodes), width)
	return &PutStreamClientService{
		filename: filepath.Base(fp),
		srcPath:  srcPath,
		destPath: dp,
		nodelist: nodelist,
		width:    width,
		node:     node,
		port:     port,
	}, nil
}

func (p *PutStreamClientService) SetFileInfo(uid, gid, filemod uint32, modtime int64) {
	p.uid = uid
	p.gid = gid
	p.filemod = filemod
	p.modtime = modtime
}

func (p *PutStreamClientService) GetSrcPath() string {
	return p.srcPath
}

func (p *PutStreamClientService) GetDestPath() string {
	return p.destPath
}

func (p *PutStreamClientService) GetPort() string {
	return p.port
}

func (p *PutStreamClientService) GetNodelist() string {
	return p.nodelist
}

func (p *PutStreamClientService) GetNodes() []string {
	return utils.ExpNodes(p.nodelist)
}

func (p *PutStreamClientService) GetAllNodelist() string {
	if p.nodelist != "" {
		return fmt.Sprintf("%s,%s", p.node, p.nodelist)
	} else {
		return p.node
	}
}

func (p *PutStreamClientService) GetWidth() int32 {
	return p.width
}

func (p *PutStreamClientService) GetStream() pb.RpcService_PutStreamClient {
	return p.stream
}

func (p *PutStreamClientService) GenStreamWithContext(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%s", p.node, p.port)
	// conn, err := grpc.Dial(addr, grpc.WithInsecure())
	ctx1, _ := context.WithTimeout(context.Background(), time.Second*1)
	conn, err := grpc.DialContext(ctx1, addr, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		return utils.GrpcErrorWrapper(err)
	}
	log.Debugf("Dial Server [%s]\n", addr)
	client := pb.NewRpcServiceClient(conn)
	stream, err := client.PutStream(ctx)
	if err != nil {
		return utils.GrpcErrorWrapper(err)
	}

	log.Debugf("Connect Server [%s]\n", addr)
	p.stream = stream
	return nil
}

func (p *PutStreamClientService) Send(data []byte) error {
	putStreamReq := &pb.PutStreamReq{
		Name:     p.filename,
		Md5:      utils.Md5sum(data),
		Location: p.destPath,
		Body:     data,
		Sn:       utils.Hostname(),
		Nodelist: p.nodelist,
		Port:     p.port,
		Width:    p.width,
		Uid:      p.uid,
		Gid:      p.gid,
		Filemod:  p.filemod,
		Modtime:  p.modtime,
	}
	return p.stream.Send(putStreamReq)
}

func (p *PutStreamClientService) CloseAndRecv() (*pb.PutStreamResp, error) {
	replay, err := p.stream.CloseAndRecv()
	return replay, utils.GrpcErrorWrapper(err)
}

func (p *PutStreamClientService) RunServe(ctx context.Context, buffer int) error {
	fp, err := os.Open(p.srcPath)
	if err != nil {
		return err
	}
	fi, err := fp.Stat()
	if err != nil {
		return err
	}
	p.modtime = fi.ModTime().Unix()
	p.filemod = uint32(fi.Mode().Perm())
	p.uid = fi.Sys().(*syscall.Stat_t).Uid
	p.gid = fi.Sys().(*syscall.Stat_t).Gid
	cnt := 0
	log.Debug("Client Stream Serve ... ")
LOOP:
	for {
		select {
		case <-ctx.Done():
			log.Debugf("Cancel Client, cnt=[%d]\n", cnt)
			break LOOP
		default:
			bufferBytes := make([]byte, buffer)
			n, err := fp.Read(bufferBytes)
			if err == io.EOF && n == 0 {
				log.Debugf("Read FILE EOF, cnt=[%d]\n", cnt)
				break LOOP
			}
			if err != nil {
				return err
			}
			data := bufferBytes[:buffer]
			err = p.Send(data)
			if err != nil {
				return err
			}
			log.Debugf("blockSize=[%d], cnt=[%d] md5=[%s]\n", buffer, cnt, utils.Md5sum(data))
		}
		cnt++
	}
	return nil
}

func (p *PutStreamClientService) Gather(replay []*pb.Replay) {
	gather(replay)
}

func PutStreamClientServiceSetup(ctx context.Context, cancel func(), localFile, destDir, nodes, buffer string, port, width int) {
	defer cancel()
	bufferSize, err := utils.ConvertSize(buffer)
	if err != nil {
		log.Error(err)
		return
	}
	clientService, err := NewPutStreamClientService(localFile, destDir, nodes, strconv.Itoa(port), int32(width))
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
		log.Errorf("PutStreamClientService Failed, err=[%s], T=[%#s]\n", err.Error(), utils.GrpcErrorMsg(err))
		// 取消或者发送失败需要汇总错误信息
		// return
	}
	log.Debug("PutStreamClientService Start Recv All Replay...")
	replays, err := clientService.CloseAndRecv()
	if err != nil {
		log.Errorf("PutStreamClientService Failed, err=[%s]\n", err.Error())
		return
	}

	clientService.Gather(replays.Replay)
	log.Debug("PutStreamClientService Stop")
}
