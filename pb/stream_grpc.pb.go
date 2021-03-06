// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.17.3
// source: pb/stream.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// RpcServiceClient is the client API for RpcService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RpcServiceClient interface {
	PutStream(ctx context.Context, opts ...grpc.CallOption) (RpcService_PutStreamClient, error)
	RunCmd(ctx context.Context, in *CmdReq, opts ...grpc.CallOption) (RpcService_RunCmdClient, error)
	Ping(ctx context.Context, in *CommonReq, opts ...grpc.CallOption) (*CommonResp, error)
}

type rpcServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewRpcServiceClient(cc grpc.ClientConnInterface) RpcServiceClient {
	return &rpcServiceClient{cc}
}

func (c *rpcServiceClient) PutStream(ctx context.Context, opts ...grpc.CallOption) (RpcService_PutStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &RpcService_ServiceDesc.Streams[0], "/pb.RpcService/PutStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &rpcServicePutStreamClient{stream}
	return x, nil
}

type RpcService_PutStreamClient interface {
	Send(*PutStreamReq) error
	Recv() (*Reply, error)
	grpc.ClientStream
}

type rpcServicePutStreamClient struct {
	grpc.ClientStream
}

func (x *rpcServicePutStreamClient) Send(m *PutStreamReq) error {
	return x.ClientStream.SendMsg(m)
}

func (x *rpcServicePutStreamClient) Recv() (*Reply, error) {
	m := new(Reply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *rpcServiceClient) RunCmd(ctx context.Context, in *CmdReq, opts ...grpc.CallOption) (RpcService_RunCmdClient, error) {
	stream, err := c.cc.NewStream(ctx, &RpcService_ServiceDesc.Streams[1], "/pb.RpcService/RunCmd", opts...)
	if err != nil {
		return nil, err
	}
	x := &rpcServiceRunCmdClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type RpcService_RunCmdClient interface {
	Recv() (*Reply, error)
	grpc.ClientStream
}

type rpcServiceRunCmdClient struct {
	grpc.ClientStream
}

func (x *rpcServiceRunCmdClient) Recv() (*Reply, error) {
	m := new(Reply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *rpcServiceClient) Ping(ctx context.Context, in *CommonReq, opts ...grpc.CallOption) (*CommonResp, error) {
	out := new(CommonResp)
	err := c.cc.Invoke(ctx, "/pb.RpcService/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RpcServiceServer is the server API for RpcService service.
// All implementations should embed UnimplementedRpcServiceServer
// for forward compatibility
type RpcServiceServer interface {
	PutStream(RpcService_PutStreamServer) error
	RunCmd(*CmdReq, RpcService_RunCmdServer) error
	Ping(context.Context, *CommonReq) (*CommonResp, error)
}

// UnimplementedRpcServiceServer should be embedded to have forward compatible implementations.
type UnimplementedRpcServiceServer struct {
}

func (UnimplementedRpcServiceServer) PutStream(RpcService_PutStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method PutStream not implemented")
}
func (UnimplementedRpcServiceServer) RunCmd(*CmdReq, RpcService_RunCmdServer) error {
	return status.Errorf(codes.Unimplemented, "method RunCmd not implemented")
}
func (UnimplementedRpcServiceServer) Ping(context.Context, *CommonReq) (*CommonResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}

// UnsafeRpcServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RpcServiceServer will
// result in compilation errors.
type UnsafeRpcServiceServer interface {
	mustEmbedUnimplementedRpcServiceServer()
}

func RegisterRpcServiceServer(s grpc.ServiceRegistrar, srv RpcServiceServer) {
	s.RegisterService(&RpcService_ServiceDesc, srv)
}

func _RpcService_PutStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(RpcServiceServer).PutStream(&rpcServicePutStreamServer{stream})
}

type RpcService_PutStreamServer interface {
	Send(*Reply) error
	Recv() (*PutStreamReq, error)
	grpc.ServerStream
}

type rpcServicePutStreamServer struct {
	grpc.ServerStream
}

func (x *rpcServicePutStreamServer) Send(m *Reply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *rpcServicePutStreamServer) Recv() (*PutStreamReq, error) {
	m := new(PutStreamReq)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _RpcService_RunCmd_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(CmdReq)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(RpcServiceServer).RunCmd(m, &rpcServiceRunCmdServer{stream})
}

type RpcService_RunCmdServer interface {
	Send(*Reply) error
	grpc.ServerStream
}

type rpcServiceRunCmdServer struct {
	grpc.ServerStream
}

func (x *rpcServiceRunCmdServer) Send(m *Reply) error {
	return x.ServerStream.SendMsg(m)
}

func _RpcService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CommonReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RpcServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.RpcService/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RpcServiceServer).Ping(ctx, req.(*CommonReq))
	}
	return interceptor(ctx, in, info, handler)
}

// RpcService_ServiceDesc is the grpc.ServiceDesc for RpcService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RpcService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pb.RpcService",
	HandlerType: (*RpcServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _RpcService_Ping_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "PutStream",
			Handler:       _RpcService_PutStream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "RunCmd",
			Handler:       _RpcService_RunCmd_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "pb/stream.proto",
}
