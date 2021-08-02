package service

import (
	"context"
	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"
)

func (p *putStreamServer) Ping(ctx context.Context, gg *pb.GG) (*pb.GG, error) {
	log.Debugf("Service Ping From [%s]\n", gg.GetHH())
	return &pb.GG{HH: utils.Hostname()}, nil
}
