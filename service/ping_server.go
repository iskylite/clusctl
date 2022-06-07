package service

import (
	"context"
	"myclush/global"
	log "myclush/logger"
	"myclush/pb"
)

func (p *putStreamServer) Ping(ctx context.Context, req *pb.CommonReq) (*pb.CommonResp, error) {
	if req.GetVersion() == global.VERSION {
		return &pb.CommonResp{Ok: true}, nil
	}
	log.Errorf("[ping] server: %s, client: %s\n", global.VERSION, req.GetVersion())
	return &pb.CommonResp{Ok: false}, nil
}
