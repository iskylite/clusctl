package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"myclush/logger"
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
	conn     *grpc.ClientConn
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
	return &PutStreamClientService{
		filename: filepath.Base(fp),
		srcPath:  srcPath,
		destPath: dp,
		nodelist: nodelist,
		width:    width,
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

func (p *PutStreamClientService) CloseConn() {
	logger.Debugf("close conn %s\n", p.node)
	p.conn.Close()
}

func (p *PutStreamClientService) checkConn(ctx context.Context, node string) (*grpc.ClientConn, pb.RpcService_PutStreamClient, error) {
	var waitc chan struct{} = make(chan struct{})
	var conn *grpc.ClientConn
	var stream pb.RpcService_PutStreamClient
	var err error
	tctx, tcancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer tcancel()
	go func() {
		defer close(waitc)
		addr := fmt.Sprintf("%s:%s", node, p.port)
		conn, err = grpc.DialContext(ctx, addr, grpc.WithInsecure())
		if err != nil {
			logger.Error(err)
			return
		}
		client := pb.NewRpcServiceClient(conn)
		stream, err = client.PutStream(ctx)
		if err != nil {
			// logger.Error(err)
			return
		}
		log.Debugf("Gen client stream -> %s\n", addr)
	}()
	select {
	case <-tctx.Done():
		logger.Debugf("connect timeout for %s\n", node)
		return nil, nil, errors.New("timeout")
	case <-waitc:
		return conn, stream, err
	}
}

func (p *PutStreamClientService) GenStreamWithContext(ctx context.Context) ([]string, error) {
	nodes := utils.ExpNodes(p.nodelist)
	nodesNum := len(nodes)
	down := make([]string, 0)
	var conn *grpc.ClientConn
	var stream pb.RpcService_PutStreamClient
	var err error
	for i := 0; i < nodesNum; i++ {
		node := nodes[i]
		conn, stream, err = p.checkConn(ctx, node)
		if err != nil {
			down = append(down, node)
			continue
		}
		p.node, p.conn, p.stream = node, conn, stream
		p.nodelist = utils.ConvertNodelist(nodes[i+1 : nodesNum])
		break
	}
	// 只要有一个连接成功，那么err就会被赋值为nil，否则则是连接失败的错误
	// 故当err为错误的时候，所有节点都连接失败
	return down, err
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
	counts := int64(math.Ceil(float64(fi.Size()) / float64(int64(buffer))))
	cnt := 0
	log.Debug("Client Stream Serve ... ")
LOOP:
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("\rData Transmission: %d/%d\n", cnt, counts)
			log.Debugf("\nCancel Client, cnt=[%d]\n", cnt)
			break LOOP
		default:
			bufferBytes := make([]byte, buffer)
			n, err := fp.Read(bufferBytes)
			if err == io.EOF && n == 0 {
				fmt.Printf("\rData Transmission: %d/%d\n", cnt, counts)
				log.Debugf("\nRead FILE EOF, cnt=[%d]\n", cnt)
				break LOOP
			}
			if err != nil {
				fmt.Printf("\rData Transmission: %d/%d\n", cnt, counts)
				return err
			}
			cnt++
			data := bufferBytes[:buffer]
			err = p.Send(data)
			if err != nil {
				fmt.Printf("\rData Transmission: %d/%d\n", cnt, counts)
				return err
			}
			fmt.Printf("\rData Transmission: %d/%d", cnt, counts)
		}
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
	down, err := clientService.GenStreamWithContext(ctx)
	if err != nil {
		log.Errorf("PutStreamClientService Failed, err=[%s]\n",
			status.Code(err).String())
		return
	}
	err = clientService.RunServe(ctx, bufferSize)
	if err != nil {
		log.Errorf("PutStreamClientService Failed, err=[%s]\n", err.Error())
		// 取消或者发送失败需要汇总错误信息
		// return
	}
	log.Debug("PutStreamClientService Start Recv All Replay...")
	replays, err := clientService.CloseAndRecv()
	if err != nil {
		log.Errorf("PutStreamClientService Failed, err=[%s]\n", err.Error())
		return
	}
	if len(down) > 0 {
		replays.Replay = append(replays.Replay, newReplay(false, "connect failed", utils.ConvertNodelist(down)))
	}
	clientService.Gather(replays.Replay)
	log.Debug("PutStreamClientService Stop")
}
