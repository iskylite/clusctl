package service

import (
	"context"
	"errors"
	"fmt"
	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"
	"regexp"
)

func (p *putStreamServer) GetNcid(ctx context.Context, gg *pb.GG) (*pb.Replay, error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, utils.Timeout(3))
	defer cancel()
	ncid, err := getLocalNcidWithContext(ctx)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	log.Infof("Local Ncid Is [%s]\n", ncid)
	return newReplay(true, ncid, gg.GetHH()), nil
}

func getLocalNcidWithContext(ctx context.Context) (string, error) {
	if !utils.Isfile("/usr/local/glex/utils/zni_read_reg") {
		return "", fmt.Errorf("/usr/local/glex/utils/zni_read_reg not exist")
	}
	out, ok := utils.ExecuteShellCmdWithContext(ctx, "/usr/local/glex/utils/zni_read_reg -r 0x800")
	if !ok {
		return "", errors.New(out)
	}
	re := regexp.MustCompile(`.*RM_LOCAL_ID\(0x800\)\s+=\s+(0x\w+).*`)
	matches := re.FindAllStringSubmatch(string(out), -1)
	if len(matches) == 0 {
		return "", nil
	}
	ncid := matches[0][1]
	if ncid == "" {
		return "", fmt.Errorf("ncid is none on [%s]", utils.Hostname())
	}
	return ncid, nil
}

// func getLocalNcidWithTimeout(ctx context.Context, timeout int) (string, error) {
// 	ctx1, cancel := context.WithTimeout(ctx, time.Duration(timeout))
// 	defer cancel()
// 	return getLocalNcidWithContext(ctx1)
// }
