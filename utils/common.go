package utils

import (
	"crypto/md5"
	"fmt"
	"myclush/pb"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-basic/uuid"
)

func Hostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	return hostname
}

func Md5sum(bytes []byte) string {
	return fmt.Sprintf("%x", md5.Sum(bytes))
}

func Isfile(fp string) bool {
	f, err := os.Stat(fp)
	if os.IsNotExist(err) {
		return false
	}
	return !f.IsDir()
}

func IsDir(fp string) bool {
	f, err := os.Stat(fp)
	if os.IsNotExist(err) {
		return false
	}
	return f.IsDir()
}

func UUID() string {
	return uuid.New()
}

func Timeout(timeout int) time.Duration {
	return time.Second * time.Duration(timeout)
}

func ConvertSize(size string) (int, error) {
	re := regexp.MustCompile(`^(\d+)([kKmM]?)$`)
	matches := re.FindAllStringSubmatch(size, -1)
	if len(matches) == 0 || len(matches[0]) != 3 {
		return 0, fmt.Errorf("block size syntax error")
	}
	blockSizeStr := matches[0][1]
	blockSize, err := strconv.Atoi(blockSizeStr)
	if err != nil {
		return 0, err
	}
	unit := strings.ToLower(matches[0][2])
	switch unit {
	case "m":
		blockSize *= 1024 * 1024
	case "k":
		blockSize *= 1024
	case "":

	default:
		return 0, fmt.Errorf("block size syntax error")
	}
	return blockSize, nil
}

func DataAggregation(replay []*pb.Replay) (sync.Map, sync.Map) {
	dataPassMap := sync.Map{}
	dataFailMap := sync.Map{}
	replayChan := make(chan *pb.Replay, runtime.NumCPU())
	go func() {
		defer close(replayChan)
		for _, r := range replay {
			replayChan <- r
		}
	}()
	for r := range replayChan {
		if r.Pass {
			nodes, ok := dataPassMap.Load(r.Msg)
			if ok {
				nodes, ok := (nodes).([]string)
				if !ok {
					dataPassMap.Store(r.Msg, ExpNodes(r.Nodelist))
				}
				dataPassMap.Store(r.Msg, append(nodes, ExpNodes(r.Nodelist)...))
			} else {
				dataPassMap.Store(r.Msg, ExpNodes(r.Nodelist))
			}
		} else {
			nodes, ok := dataFailMap.Load(r.Msg)
			if ok {
				nodes, ok := (nodes).([]string)
				if !ok {
					dataFailMap.Store(r.Msg, ExpNodes(r.Nodelist))
				}
				dataFailMap.Store(r.Msg, append(nodes, ExpNodes(r.Nodelist)...))
			} else {
				dataFailMap.Store(r.Msg, ExpNodes(r.Nodelist))
			}
		}
	}
	return dataPassMap, dataFailMap
}
