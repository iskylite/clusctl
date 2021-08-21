package service

import (
	"context"
	"errors"
	"fmt"
	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"google.golang.org/grpc"
)

var LocalNode string = utils.Hostname()

type putStreamServer struct {
	tmpDir     string
	grpcServer *grpc.Server
}

func NewPutStreamServerService(tmpDir string) (*putStreamServer, error) {
	// 本地默认临时目录
	tmpDir = filepath.Join("/tmp", tmpDir)
	if !utils.IsDir(tmpDir) {
		err := os.Mkdir(tmpDir, 0644)
		if err != nil {
			return nil, err
		}
	}
	return &putStreamServer{
		tmpDir: tmpDir,
	}, nil
}

func clearTempDir(temp string) {
	if err := os.RemoveAll(temp); err != nil {
		log.Errorf("clear temp error: %s\n", err)
		return
	}
	log.Debug("clear temp dir  Before server stop")
}

func (p *putStreamServer) RunServer(port string) error {
	if port == "" {
		return errors.New("server port not been specified")
	}
	addr := fmt.Sprintf("0.0.0.0:%s", port)
	conn, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	p.grpcServer = grpc.NewServer()
	pb.RegisterRpcServiceServer(p.grpcServer, p)
	err = p.grpcServer.Serve(conn)
	if err != nil {
		return err
	}
	return nil
}

func (p *putStreamServer) Stop() {
	if p.grpcServer == nil {
		return
	}
	p.grpcServer.Stop()
}

func PutStreamServerServiceSetup(ctx context.Context, cancel func(), tmpDir string, port int) {
	serverService, err := NewPutStreamServerService(tmpDir)
	if err != nil {
		log.Error(err)
		return
	}
	defer clearTempDir(serverService.tmpDir)
	go func() {
		defer cancel()
		err := serverService.RunServer(strconv.Itoa(port))
		if err != nil {
			log.Errorf("PutStreamServerService Failed, err=[%s]\n", utils.GrpcErrorMsg(err))
			return
		}
	}()
	<-ctx.Done()
	serverService.Stop()
	log.Info("PutStreamServerService Stop")
}
