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
	"syscall"

	"google.golang.org/grpc"
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
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	// ctx1, cel := context.WithTimeout(context.Background(), time.Second*3)
	// defer cel()
	// conn, err := grpc.DialContext(ctx1, addr, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Debugf("Dial Server [%s]\n", addr)
	client := pb.NewRpcServiceClient(conn)
	stream, err := client.PutStream(ctx)
	if err != nil {
		log.Error(err.Error())
		return err
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
	return replay, err
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
	log.Debug("Client Stream Serve")
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
				log.Error(err.Error())
				return err
			}
			log.Debugf("Send Stream: cnt=[%d] md5=[%s]\n", cnt, utils.Md5sum(data))
		}
		cnt++
	}
	return nil
}

func (p *PutStreamClientService) Gather(replay []*pb.Replay) (idleNodes, downNodes, cancelNodes []string) {
	idleNodes = make([]string, 0)
	downNodes = make([]string, 0)
	cancelNodes = make([]string, 0)
	for _, rep := range replay {
		nodelist := utils.ExpNodes(rep.Nodelist)
		if rep.Pass {
			idleNodes = append(idleNodes, nodelist...)
		} else {
			if rep.Msg == "canceled" {
				cancelNodes = append(cancelNodes, nodelist...)
				log.Debugf("nodelist=%s, error=canceled\n", rep.Nodelist)
			} else {
				downNodes = append(downNodes, nodelist...)
				log.Debugf("nodelist=%s, error=%s\n", rep.Nodelist, rep.Msg)
			}
		}
	}
	return
}
