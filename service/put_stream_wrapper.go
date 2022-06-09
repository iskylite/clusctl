package service

import "myclush/pb"

type Wrapper interface {
	// 如果传输出错或者grpc被cancel，此时需要设置不在接受数据
	SetBad()
	// 直接发送数据
	Send([]byte) error
	// 对Send的封装，从管道中接收数据并发送
	SendFromChannel()
	// 注册响应管道
	DiscribeRepliesChannel(repliesChannel chan *pb.Reply)
	// 获取本次连接关联的所有子节点
	GetAllNodelist() string
	// 接受数据，发送到管道
	RecvData(data []byte)
	// 获取分发节点的主节点
	GetBatchNode() string
	// 关闭数据接受管道
	CloseDataChan()
	// 关闭连接
	CloseConn()
	// 是否时本地连接（不是远程grpc连接）
	IsLocal() bool
	// 清理通道中的数据
	CleanDataChan()
}
