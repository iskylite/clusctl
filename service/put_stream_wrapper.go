package service

import (
	"context"
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

type Wrapper interface {
	SetBad()
	Send([]byte) error
	SendFromChannel()
	DiscribeRepliesChannel(repliesChannel chan *pb.Reply)
	GetAllNodelist() string
	RecvData(data []byte)
	GetBatchNode() string
	CloseDataChan()
	CloseConn()
	IsLocal() bool
}

type LocalWriterWrapper struct {
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
	r := new(LocalWriterWrapper)
	var ok atomic.Value
	ok.Store(true)
	r.fp, r.f, r.dataChan, r.Ok, r.wg, r.ctx = fp, f, make(chan []byte, runtime.NumCPU()), ok, wg, ctx
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
	defer log.Debugf("stop LocalStreamClientService %s\n", l.GetBatchNode())
	defer l.wg.Done()
	defer func() {
		err := l.CloseAndRecv()
		if err != nil {
			l.repliesChannel <- newReply(false, err.Error(), LocalNode)
		} else {
			l.repliesChannel <- newReply(true, "Success", LocalNode)
		}
	}()
	cnt := 0
LOOP:
	for {
		select {
		case <-l.ctx.Done():
			l.SetBad()
			log.Errorf("cnt=[%d], canceled\n", cnt)
			l.err = status.Error(codes.Canceled, "context canceled")
			break LOOP
		case data, ok := <-l.dataChan:
			if !ok {
				log.Debugf("cnt=[%d], LOCAL DATA EOF\n", cnt)
				break LOOP
			}
			err := l.Send(data)
			if err != nil {
				log.Errorf("cnt=[%d], %v\n", cnt, err)
				l.err = err
				l.SetBad()
				break LOOP
			}
			cnt++
		}
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
	} else {
		if err := l.f.Close(); err != nil {
			log.Error(err)
		}
		return l.err
	}
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
	close(l.dataChan)
}
