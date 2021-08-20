package service

import (
	"context"
	"myclush/pb"
	"myclush/utils"
	"sync"

	// log "myclush/logger"
	"runtime"
)

type StreamWrapper struct {
	Ok       bool
	stream   *PutStreamClientService
	dataChan chan []byte
	wg       *sync.WaitGroup
	replay   *pb.PutStreamResp
	err      error
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
	return &StreamWrapper{true, stream, make(chan []byte, runtime.NumCPU()), wg, nil, nil}, down, nil
}

func (s *StreamWrapper) SetFileInfo(uid, gid, filemod uint32, modtime int64) {
	s.stream.SetFileInfo(uid, gid, filemod, modtime)
}

func (s *StreamWrapper) SetBad() {
	s.Ok = false
}

func (s *StreamWrapper) Send(body []byte) error {
	if s.Ok {
		return s.stream.Send(body)
	}
	return nil
}

func (s *StreamWrapper) SendFromChannel() {
	defer s.wg.Done()
	for data := range s.dataChan {
		err := s.Send(data)
		if err != nil {
			s.SetBad()
			break
		}
	}
	s.replay, s.err = s.CloseAndRecv()
}

func (s *StreamWrapper) CloseAndRecv() (*pb.PutStreamResp, error) {
	if !s.Ok {
		return nil, nil
	}
	return s.stream.CloseAndRecv()
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

func (s *StreamWrapper) GetDataChan() chan []byte {
	return s.dataChan
}

func (s *StreamWrapper) CloseDataChan() {
	close(s.dataChan)
}

func (s *StreamWrapper) GetResult() (*pb.PutStreamResp, error) {
	return s.replay, s.err
}

func (s *StreamWrapper) CloseConn() {
	s.stream.CloseConn()
}

func (s *StreamWrapper) IsLocal() bool {
	return false
}
