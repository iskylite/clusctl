package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LocalWriterWrapper struct {
	node           string
	fp             string
	f              *os.File
	dataChan       chan []byte
	repliesChannel chan *pb.Reply
	Ok             atomic.Value
	wg             *sync.WaitGroup
	ctx            context.Context
	err            error
}

func newLocalWriterWrapper(ctx context.Context, data *pb.PutStreamReq, tmpDir string, wg *sync.WaitGroup) (*LocalWriterWrapper, error) {
	// 创建临时文件
	f, err := os.CreateTemp(tmpDir, utils.UUID())
	if err != nil {
		return nil, err
	}
	tmpfile := f.Name()
	// 修改文件权限
	// 属组
	if err = f.Chown(int(data.Uid), int(data.Gid)); err != nil {
		return nil, err
	}
	// 权限
	if err = f.Chmod(os.FileMode(data.Filemod)); err != nil {
		return nil, err
	}
	// 修改时间
	if err = os.Chtimes(tmpfile, time.Now(), time.Unix(data.Modtime, 0)); err != nil {
		return nil, err
	}
	fp := filepath.Join(data.Location, data.Name)
	if !utils.IsDir(data.Location) {
		fp = data.Location
	}
	r := new(LocalWriterWrapper)
	r.node = data.GetNode()
	var ok atomic.Value
	ok.Store(true)
	r.fp, r.f, r.dataChan, r.Ok, r.wg, r.ctx = fp, f, make(chan []byte, runtime.NumCPU()/2), ok, wg, ctx
	return r, nil
}

func (l *LocalWriterWrapper) DiscribeRepliesChannel(repliesChannel chan *pb.Reply) {
	l.repliesChannel = repliesChannel
}

func (l *LocalWriterWrapper) SetBad() {
	l.Ok.Store(false)
}

func (l *LocalWriterWrapper) Send(data []byte) error {
	if l.Ok.Load().(bool) {
		_, err := l.f.Write(data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *LocalWriterWrapper) CloseConn() {
	log.Debugf("close conn local\n")
}

func (l *LocalWriterWrapper) IsLocal() bool {
	return true
}

func (l *LocalWriterWrapper) SendFromChannel() {
	defer log.Debugf("stop LocalStreamClientService in %s\n", l.GetBatchNode())
	defer l.wg.Done()
	cnt := 0
	isCanceled := false
LOOP:
	for {
		select {
		case <-l.ctx.Done():
			l.SetBad()
			errMsg := fmt.Sprintf("break [%d], cancel from context\n", cnt)
			log.Error(errMsg)
			l.err = status.Error(codes.Canceled, "errMsg")
			isCanceled = true
			break LOOP
		case data, ok := <-l.dataChan:
			if !ok {
				if !l.Ok.Load().(bool) {
					// 非EOF
					// TODO：此处应该是无用代码
					errMsg := fmt.Sprintf("break [%d], cancel from dataChan\n", cnt)
					l.err = status.Error(codes.Canceled, errMsg)
					log.Debugf(errMsg)
					isCanceled = true
				} else {
					log.Debugf("break [%d], LOCAL DATA EOF\n", cnt)
				}
				break LOOP
			}
			err := l.Send(data)
			if err != nil {
				log.Errorf("break [%d], send error: %v\n", cnt, err)
				l.err = err
				l.SetBad()
				break LOOP
			}
			cnt++
		}
	}
	// 文件重命名和删除
	err := l.CloseAndRecv()
	if err != nil {
		log.Debug(err.Error())
		if isCanceled {
			log.Debug("comfirm cancel signal, abort reply")
			return
		}
		l.repliesChannel <- newReply(false, err.Error(), l.node)
	} else {
		l.repliesChannel <- newReply(true, "Success", l.node)
	}
}

func (l *LocalWriterWrapper) CloseAndRecv() error {
	defer os.Remove(l.f.Name())
	if l.err == nil {
		if err := l.f.Close(); err != nil {
			return err
		}
		if err := os.Rename(l.f.Name(), l.fp); err != nil {
			return err
		}
		log.Debugf("rename %s to %s\n", l.f.Name(), l.fp)
		return nil
	}
	if err := l.f.Close(); err != nil {
		log.Error(err)
	}
	return l.err
}

func (l *LocalWriterWrapper) GetAllNodelist() string {
	return utils.Hostname()
}

func (l *LocalWriterWrapper) GetBatchNode() string {
	return utils.Hostname()
}

func (l *LocalWriterWrapper) RecvData(data []byte) {
	if l.Ok.Load().(bool) {
		l.dataChan <- data
	}
}

func (l *LocalWriterWrapper) CloseDataChan() {
	log.Infof("close dataChan to [%s]\n", l.node)
	close(l.dataChan)
}

// CleanDataChan clear all data in dataChan
func (l *LocalWriterWrapper) CleanDataChan() {
	log.Infof("clear dataChan to [%s]\n", l.node)
	for range l.dataChan {
	}
}
