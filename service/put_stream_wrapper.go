package service

import (
	"context"
	"errors"
	"os"
	"runtime"
	"sync"

	"myclush/logger"
	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"
)

func newReplay(pass bool, msg, nodeList string) *pb.Replay {
	return &pb.Replay{
		Pass:     pass,
		Msg:      msg,
		Nodelist: nodeList,
	}
}

func newPutStreamResp(replay []*pb.Replay) *pb.PutStreamResp {
	return &pb.PutStreamResp{Replay: replay}
}

type Wrapper interface {
	SetBad()
	Send([]byte) error
	SendFromChannel()
	CloseAndRecv() (*pb.PutStreamResp, error)
	GetAllNodelist() string
	GetDataChan() chan []byte
	GetBatchNode() string
	CloseDataChan()
	GetResult() (*pb.PutStreamResp, error)
	CloseConn()
	IsLocal() bool
}

type LocalWriterWrapper struct {
	fp       string
	f        *os.File
	dataChan chan []byte
	Ok       bool
	wg       *sync.WaitGroup
	ctx      context.Context
	replay   *pb.PutStreamResp
	err      error
}

func newLocalWriterWrapper(ctx context.Context, fp string, f *os.File, wg *sync.WaitGroup) (*LocalWriterWrapper, error) {
	return &LocalWriterWrapper{fp, f, make(chan []byte, runtime.NumCPU()), true, wg, ctx, nil, nil}, nil
}

func (l *LocalWriterWrapper) SetBad() {
	l.Ok = false
}

func (l *LocalWriterWrapper) Send(data []byte) error {
	if l.Ok {
		_, err := l.f.Write(data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *LocalWriterWrapper) CloseConn() {
	logger.Debugf("close conn local\n")
}

func (l *LocalWriterWrapper) IsLocal() bool {
	return true
}

func (l *LocalWriterWrapper) SendFromChannel() {
	defer l.wg.Done()
	cnt := 0
LOOP:
	for {
		select {
		case <-l.ctx.Done():
			log.Errorf("Write Data Into LocalFile Canceled, cnt=[%d]\n", cnt)
			l.err = errors.New("canceled")
			break LOOP
		case data, ok := <-l.dataChan:
			if !ok {
				log.Debugf("Write Data EOF Into LocalFile, cnt=[%d]\n", cnt)
				break LOOP
			}
			err := l.Send(data)
			if err != nil {
				log.Error(err)
				l.err = err
				l.SetBad()
				break LOOP
			}
			log.Debugf("Write Data Into LocalFile, cnt=[%d]\n", cnt)
			cnt++
		}
	}
	l.replay, l.err = l.CloseAndRecv()
}

func (l *LocalWriterWrapper) CloseAndRecv() (*pb.PutStreamResp, error) {
	defer l.f.Close()
	if l.err != nil {
		if err := os.Remove(l.f.Name()); err != nil {
			return nil, err
		}
		return nil, l.err
	}
	err := os.Rename(l.f.Name(), l.fp)
	if err != nil {
		return nil, err
	}
	return newPutStreamResp([]*pb.Replay{newReplay(true, "", l.GetAllNodelist())}), nil
}

func (l *LocalWriterWrapper) GetAllNodelist() string {
	return utils.Hostname()
}

func (l *LocalWriterWrapper) GetBatchNode() string {
	return utils.Hostname()
}

func (l *LocalWriterWrapper) GetDataChan() chan []byte {
	return l.dataChan
}

func (l *LocalWriterWrapper) CloseDataChan() {
	close(l.dataChan)
}

func (l *LocalWriterWrapper) GetResult() (*pb.PutStreamResp, error) {
	return l.replay, l.err
}
