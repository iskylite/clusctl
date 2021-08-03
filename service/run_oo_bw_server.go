package service

import (
	"context"
	"fmt"
	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"
	"path/filepath"
)

func (p *putStreamServer) RunOoBwServer(ctx context.Context, req *pb.OoBwServerReq) (*pb.Replay, error) {
	var cmdFile, args string
	if req.Length != 0 && req.Count != 0 {
		cmdFile = "/usr/local/glex/examples/oo_bw_r_loop"
		args = fmt.Sprintf("8 %s 8 %s %d %d", req.Ncid, req.Buffer, req.Length, req.Count)
	} else {
		cmdFile = "/usr/local/glex/examples/oo_bw_r"
		args = fmt.Sprintf("8 %s 8 %s", req.Ncid, req.Buffer)
	}
	if !utils.Isfile(cmdFile) {
		log.Errorf("%s not exist", cmdFile)
		return nil, fmt.Errorf("%s not exist", cmdFile)
	}
	execname := filepath.Base(cmdFile)
	command := fmt.Sprintf("%s %s", cmdFile, args)
	log.Debugf("%s start on %s\n", execname, utils.LocalTime())
	out, err := utils.ExecuteShellCmdWithContext(ctx, command)
	defer log.Debugf("%s finish on %s\n", execname, utils.LocalTime())
	if err != nil {
		log.Error(err)
		log.Error(out)
		return nil, err
	}
	return newReplay(true, string(out), utils.Hostname()), nil
}
