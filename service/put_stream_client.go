package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math"
	"myclush/global"
	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/iskylite/nodeset"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PutStreamClientService struct {
	filename string
	srcPath  string
	destPath string
	port     string
	nodelist string
	node     string
	num      int
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

// 获取节点树宽度
func (p *PutStreamClientService) GetWidth() int32 {
	return p.width
}

// 获取grpc数据流
func (p *PutStreamClientService) GetStream() pb.RpcService_PutStreamClient {
	return p.stream
}

func (p *PutStreamClientService) CloseConn() {
	log.Debugf("close conn %s\n", p.node)
	p.conn.Close()
}

// 检查到目标节点的连接是否正常可用
func (p *PutStreamClientService) checkConn(ctx context.Context, node string, authority grpc.DialOption) (*grpc.ClientConn, pb.RpcService_PutStreamClient, error) {
	var waitc chan struct{} = make(chan struct{})
	var conn *grpc.ClientConn
	var stream pb.RpcService_PutStreamClient
	var err error
	tctx, tcancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer tcancel()
	go func() {
		defer close(waitc)
		addr := fmt.Sprintf("%s:%s", node, p.port)
		conn, err = grpc.DialContext(ctx, addr, authority, global.ClientTransportCredentials)
		if err != nil {
			log.Error(err)
			return
		}
		client := pb.NewRpcServiceClient(conn)
		stream, err = client.PutStream(ctx)
		if err != nil {
			log.Error(err)
			return
		}
		log.Debugf("Gen client stream -> %s\n", addr)
	}()
	select {
	case <-tctx.Done():
		log.Errorf("connect timeout for %s\n", node)
		return nil, nil, status.Error(codes.DeadlineExceeded, "connect timeout")
	case <-waitc:
		if err != nil {
			return conn, stream, err
		}
		return conn, stream, err
	}
}

// 生成grpc流
func (p *PutStreamClientService) GenStreamWithContext(ctx context.Context, authority grpc.DialOption) ([]string, error) {
	nodes := utils.ExpNodes(p.nodelist)
	p.num = len(nodes)
	down := make([]string, 0)
	var conn *grpc.ClientConn
	var stream pb.RpcService_PutStreamClient
	var err error
	for i := 0; i < p.num; i++ {
		node := nodes[i]
		conn, stream, err = p.checkConn(ctx, node, authority)
		if err != nil {
			down = append(down, node)
			continue
		}
		p.node, p.conn, p.stream = node, conn, stream
		p.nodelist = utils.Merge(nodes[i+1 : p.num]...)
		break
	}
	// 只要有一个连接成功，那么err就会被赋值为nil，否则则是连接失败的错误
	// 故当err为错误的时候，所有节点都连接失败
	return down, err
}

// 发送数据
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

// 用于客户端，开启服务
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
			fmt.Printf("\r数据读取: %d/%d %s\n", cnt, counts, log.ColorWrapper("CANCEL", log.Cancel))
			log.Debugf("\nCancel Client, cnt=[%d]\n", cnt)
			break LOOP
		default:
			bufferBytes := make([]byte, buffer)
			n, err := fp.Read(bufferBytes)
			if err == io.EOF && n == 0 {
				fmt.Printf("\r数据读取: %d/%d %s\n", cnt, counts, log.ColorWrapper("EOF", log.Success))
				log.Debugf("Read FILE EOF, cnt=[%d]\n", cnt)
				if err = p.stream.CloseSend(); err != nil {
					log.Errorf("close send failed, %v\n", err)
				}
				break LOOP
			}
			if err != nil {
				fmt.Printf("\r数据读取: %d/%d %s %s\n", cnt, counts, log.ColorWrapper("ERROR", log.Failed), log.ColorWrapper(err.Error(), log.Failed))
				return err
			}
			cnt++
			data := bufferBytes[0:n]
			if err = p.Send(data); err != nil {
				fmt.Printf("\r数据读取: %d/%d %s %s\n", cnt, counts, log.ColorWrapper("ERROR ==>", log.Failed), log.ColorWrapper(utils.GrpcErrorMsg(err), log.Failed))
				return err
			}
			fmt.Printf("\r数据读取: %d/%d\r", cnt, counts)
		}
	}
	return nil
}

func (p *PutStreamClientService) Gather(reply []*pb.Reply) {
	gather(reply)
}

// 把节点映射到map中，方便判断是否有该节点的响应
// 20W节点0.4s运行完毕
func hashNodesMap(nodes string) (sync.Map, error) {
	var resultSet sync.Map
	iter, err := nodeset.Yield(nodes)
	if err != nil {
		return resultSet, err
	}
	maxWorkers := runtime.NumCPU()
	nodeChannel := make(chan string, maxWorkers)
	var wg sync.WaitGroup
	wg.Add(maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		go func() {
			defer wg.Done()
			for node := range nodeChannel {
				resultSet.Store(node, false)
			}
		}()
	}
	for iter.Next() {
		nodeChannel <- iter.Value()
	}
	close(nodeChannel)
	wg.Wait()
	return resultSet, nil
}

// 客户端服务
// 用于myclush
func PutStreamClientServiceSetup(ctx context.Context, cancel func(), localFile, destDir, nodes, buffer string, port, width int) {
	defer cancel()
	bufferSize, err := utils.ConvertSize(buffer)
	if err != nil {
		log.Error(err)
		return
	}
	// resultSet存储每个节点的是否有响应，用于在enter时输出当前没有拿到响应的节点列表
	resOriginMap, err := hashNodesMap(nodes)
	if err != nil {
		log.Error(err)
		return
	}
	// log.Info("hashNodesMap done")
	clientService, err := NewPutStreamClientService(localFile, destDir, nodes, strconv.Itoa(port), int32(width))
	if err != nil {
		log.Errorf("PutStreamClientService Failed, err=[%s]\n", err.Error())
		return
	}
	resps := make([]*pb.Reply, 0)
	down, err := clientService.GenStreamWithContext(ctx, global.Authority)
	if err != nil {
		log.Errorf("PutStreamClientService Failed, err=[%s]\n",
			status.Code(err).String())
		return
	}
	defer clientService.CloseConn()
	// 获取运行状态下未获取到响应的节点
	go func() {
		for {
			idle := make([]string, 0)
			stdinBuf := bufio.NewReaderSize(os.Stdin, 1)
			key, _ := stdinBuf.ReadByte()
			if key == 10 {
				// press enter
				resOriginMap.Range(func(key, value interface{}) bool {
					if !value.(bool) {
						idle = append(idle, key.(string))
					}
					return true
				})
				idleNodes := utils.Merge(idle...)
				fmt.Printf("\r等待结果: %s\n", idleNodes)
			}
		}
	}()
	cnt := 0
	downCnt := len(down)
	if downCnt > 0 {
		cnt += downCnt
		resps = append(resps, newReply(false, "connect failed", utils.Merge(down...)))
		fmt.Printf("\r结果汇总: %d/%d\r", cnt, clientService.num)
		for _, node := range down {
			resOriginMap.Store(node, true)
		}
	}
	var waitc sync.WaitGroup
	waitc.Add(1)
	go func() {
		defer waitc.Done()
	LOOP:
		for {
			data, err := clientService.stream.Recv()
			switch err {
			case nil:
				resps = append(resps, data)
				cnt += len(utils.ExpNodes(data.Nodelist))
				fmt.Printf("\r结果汇总: %d/%d", cnt, clientService.num)
				for _, node := range utils.ExpNodes(data.Nodelist) {
					resOriginMap.Store(node, true)
				}
			case io.EOF:
				fmt.Printf("\r结果汇总: %d/%d %s\n", cnt, clientService.num, log.ColorWrapper("EOF", log.Success))
				break LOOP
			default:
				fmt.Printf("\r结果汇总: %d/%d %s %s\n", cnt, clientService.num, log.ColorWrapper("ERROR ==>", log.Failed),
					log.ColorWrapper(utils.GrpcErrorMsg(err), log.Failed))
				break LOOP
			}
		}
	}()
	err = clientService.RunServe(ctx, bufferSize)
	if err != nil {
		log.Errorf("PutStreamClientService Failed, err=[%v]\n", err)
		// 取消或者发送失败需要汇总错误信息
		cancel()
		return
	}
	log.Debug("PutStreamClientService Start Recv All Replies...")
	waitc.Wait()
	clientService.Gather(resps)
	log.Debug("PutStreamClientService Stop")
}
