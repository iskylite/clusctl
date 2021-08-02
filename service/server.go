package service

import (
	"errors"
	"net"
	"fmt"
	"myclush/pb"
	
	"google.golang.org/grpc"
)

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
