package service

import (
	"context"
	"errors"
	"io"
	"myclush/logger"
	log "myclush/logger"
	"myclush/pb"
	"myclush/utils"
	"sync"
	"sync/atomic"
	"time"

	"runtime"
)

type StreamWrapper struct {
	Ok              atomic.Value
	stream          *PutStreamClientService
	dataChan        chan []byte
	replaiesChannel chan *pb.Replay
	wg              *sync.WaitGroup
	replay          *pb.PutStreamResp
	err             error
}

func newStreamWrapper(ctx context.Context, filename, destPath, port string, nodes []string, width int32, wg *sync.WaitGroup) (*StreamWrapper, []string, error) {
	nodelist := utils.ConvertNodelist(nodes)
	stream := &PutStreamClientService{
		filename: filename,
		destPath: destPath,
		nodelist: nodelist,
		port:     port,
		// node:     nodes[0],
		width: width,
	}
	down, err := stream.GenStreamWithContext(ctx)
	if err != nil {
		return nil, down, err
	}
	var ok atomic.Value
	ok.Store(true)
	return &StreamWrapper{ok, stream, make(chan []byte, runtime.NumCPU()), nil, wg, nil, nil}, down, nil
}

func (s *StreamWrapper) SetFileInfo(uid, gid, filemod uint32, modtime int64) {
	s.stream.SetFileInfo(uid, gid, filemod, modtime)
}

func (s *StreamWrapper) DiscribeReplaiesChannel(replaiesChannel chan *pb.Replay) {
	s.replaiesChannel = replaiesChannel
}

func (s *StreamWrapper) SetBad() {
	s.Ok.Store(false)
}

func (s *StreamWrapper) Send(body []byte) error {
	var err error
	if s.Ok.Load().(bool) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		var waitc chan struct{} = make(chan struct{})
		go func() {
			defer close(waitc)
			err = s.stream.Send(body)
		}()
		select {
		case <-ctx.Done():
			err = errors.New("send timeout")
		case <-waitc:

		}
	}
	return err
}

func (s *StreamWrapper) SendFromChannel() {
	defer log.Debugf("stop PutStreamClientService %s\n", s.stream.node)
	defer s.wg.Done()
	var waitc sync.WaitGroup
	waitc.Add(1)
	go func() {
		defer waitc.Done()
	LOOP:
		for {
			data, err := s.stream.stream.Recv()
			switch err {
			case nil:
				s.replaiesChannel <- data
				log.Debugf("node=%s, into replay=%s\n", data.Nodelist, data.Msg)
			case io.EOF:
				log.Debug("client service recv EOF")
				break LOOP
			default:
				log.Error(utils.GrpcErrorMsg(err))
				break LOOP
			}
		}
	}()
	for {
		data, ok := <-s.dataChan
		if !ok {
			log.Debugf("send close signal to %s\n", s.GetBatchNode())
			s.stream.stream.CloseSend()
			break
		}
		if err := s.Send(data); err != nil {
			logger.Error(err)
			s.SetBad()
			s.replaiesChannel <- newReplay(false, utils.GrpcErrorMsg(err), s.GetAllNodelist())
			break
		}
	}
	waitc.Wait()
}

func (s *StreamWrapper) GetNodelist() string {
	return s.stream.nodelist
}

func (s *StreamWrapper) GetNodes() []string {
	return s.stream.GetNodes()
}

func (s *StreamWrapper) GetAllNodelist() string {
	return s.stream.GetAllNodelist()
}

func (s *StreamWrapper) GetBatchNode() string {
	return s.stream.node
}

func (s *StreamWrapper) RecvData(data []byte) {
	if s.Ok.Load().(bool) {
		s.dataChan <- data
	}
}

func (s *StreamWrapper) CloseDataChan() {
	close(s.dataChan)
}

func (s *StreamWrapper) CloseConn() {
	s.stream.CloseConn()
}

func (s *StreamWrapper) IsLocal() bool {
	return false
}
