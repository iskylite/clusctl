package service

import (
	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"
	"os"
)

func newReply(pass bool, msg, nodeList string) *pb.Reply {
	return &pb.Reply{
		Pass:     pass,
		Msg:      msg,
		Nodelist: nodeList,
	}
}

func gather(replay []*pb.Reply) {
	dataPassMap, dataFailMap := utils.DataAggregation(replay)
	dataPassMap.Range(func(key, value interface{}) bool {
		k, ok := key.(string)
		if !ok {
			return false
		}
		v, ok := value.([]string)
		if !ok {
			return false
		}
		log.ColorWrapperInfo(log.Success, v, k)
		return true
	})
	dataFailMap.Range(func(key, value interface{}) bool {
		k, ok := key.(string)
		if !ok {
			return false
		}
		v, ok := value.([]string)
		if !ok {
			return false
		}
		log.ColorWrapperInfo(log.Failed, v, k)
		return true
	})
}

func gatherWithOutput(replay []*pb.Reply, f *os.File) {
	dataPassMap, dataFailMap := utils.DataAggregation(replay)
	dataPassMap.Range(func(key, value interface{}) bool {
		k, ok := key.(string)
		if !ok {
			return false
		}
		v, ok := value.([]string)
		if !ok {
			return false
		}
		log.MultiColorWrapperInfo(log.Success, v, k, f)
		return true
	})
	dataFailMap.Range(func(key, value interface{}) bool {
		k, ok := key.(string)
		if !ok {
			return false
		}
		v, ok := value.([]string)
		if !ok {
			return false
		}
		log.MultiColorWrapperInfo(log.Failed, v, k, f)
		return true
	})
}
