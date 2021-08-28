// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: stream.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// putStream request
type PutStreamReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name     string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Md5      string `protobuf:"bytes,2,opt,name=md5,proto3" json:"md5,omitempty"`
	Location string `protobuf:"bytes,3,opt,name=location,proto3" json:"location,omitempty"`
	Body     []byte `protobuf:"bytes,4,opt,name=body,proto3" json:"body,omitempty"`
	Sn       string `protobuf:"bytes,5,opt,name=sn,proto3" json:"sn,omitempty"`
	Port     string `protobuf:"bytes,6,opt,name=port,proto3" json:"port,omitempty"`
	Nodelist string `protobuf:"bytes,7,opt,name=nodelist,proto3" json:"nodelist,omitempty"`
	Width    int32  `protobuf:"varint,8,opt,name=width,proto3" json:"width,omitempty"`
	Uid      uint32 `protobuf:"varint,9,opt,name=uid,proto3" json:"uid,omitempty"`
	Gid      uint32 `protobuf:"varint,10,opt,name=gid,proto3" json:"gid,omitempty"`
	Filemod  uint32 `protobuf:"varint,11,opt,name=filemod,proto3" json:"filemod,omitempty"`
	Modtime  int64  `protobuf:"varint,12,opt,name=modtime,proto3" json:"modtime,omitempty"`
}

func (x *PutStreamReq) Reset() {
	*x = PutStreamReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stream_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PutStreamReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PutStreamReq) ProtoMessage() {}

func (x *PutStreamReq) ProtoReflect() protoreflect.Message {
	mi := &file_stream_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PutStreamReq.ProtoReflect.Descriptor instead.
func (*PutStreamReq) Descriptor() ([]byte, []int) {
	return file_stream_proto_rawDescGZIP(), []int{0}
}

func (x *PutStreamReq) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *PutStreamReq) GetMd5() string {
	if x != nil {
		return x.Md5
	}
	return ""
}

func (x *PutStreamReq) GetLocation() string {
	if x != nil {
		return x.Location
	}
	return ""
}

func (x *PutStreamReq) GetBody() []byte {
	if x != nil {
		return x.Body
	}
	return nil
}

func (x *PutStreamReq) GetSn() string {
	if x != nil {
		return x.Sn
	}
	return ""
}

func (x *PutStreamReq) GetPort() string {
	if x != nil {
		return x.Port
	}
	return ""
}

func (x *PutStreamReq) GetNodelist() string {
	if x != nil {
		return x.Nodelist
	}
	return ""
}

func (x *PutStreamReq) GetWidth() int32 {
	if x != nil {
		return x.Width
	}
	return 0
}

func (x *PutStreamReq) GetUid() uint32 {
	if x != nil {
		return x.Uid
	}
	return 0
}

func (x *PutStreamReq) GetGid() uint32 {
	if x != nil {
		return x.Gid
	}
	return 0
}

func (x *PutStreamReq) GetFilemod() uint32 {
	if x != nil {
		return x.Filemod
	}
	return 0
}

func (x *PutStreamReq) GetModtime() int64 {
	if x != nil {
		return x.Modtime
	}
	return 0
}

type Reply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Pass     bool   `protobuf:"varint,1,opt,name=pass,proto3" json:"pass,omitempty"`
	Nodelist string `protobuf:"bytes,2,opt,name=nodelist,proto3" json:"nodelist,omitempty"`
	Msg      string `protobuf:"bytes,3,opt,name=msg,proto3" json:"msg,omitempty"`
}

func (x *Reply) Reset() {
	*x = Reply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stream_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Reply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Reply) ProtoMessage() {}

func (x *Reply) ProtoReflect() protoreflect.Message {
	mi := &file_stream_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Reply.ProtoReflect.Descriptor instead.
func (*Reply) Descriptor() ([]byte, []int) {
	return file_stream_proto_rawDescGZIP(), []int{1}
}

func (x *Reply) GetPass() bool {
	if x != nil {
		return x.Pass
	}
	return false
}

func (x *Reply) GetNodelist() string {
	if x != nil {
		return x.Nodelist
	}
	return ""
}

func (x *Reply) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

// runcmd request
type CmdReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Cmd      string `protobuf:"bytes,1,opt,name=cmd,proto3" json:"cmd,omitempty"`
	Nodelist string `protobuf:"bytes,2,opt,name=nodelist,proto3" json:"nodelist,omitempty"`
	Width    int32  `protobuf:"varint,3,opt,name=width,proto3" json:"width,omitempty"`
	Port     string `protobuf:"bytes,4,opt,name=port,proto3" json:"port,omitempty"`
	Daemon   bool   `protobuf:"varint,5,opt,name=daemon,proto3" json:"daemon,omitempty"`
}

func (x *CmdReq) Reset() {
	*x = CmdReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stream_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CmdReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CmdReq) ProtoMessage() {}

func (x *CmdReq) ProtoReflect() protoreflect.Message {
	mi := &file_stream_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CmdReq.ProtoReflect.Descriptor instead.
func (*CmdReq) Descriptor() ([]byte, []int) {
	return file_stream_proto_rawDescGZIP(), []int{2}
}

func (x *CmdReq) GetCmd() string {
	if x != nil {
		return x.Cmd
	}
	return ""
}

func (x *CmdReq) GetNodelist() string {
	if x != nil {
		return x.Nodelist
	}
	return ""
}

func (x *CmdReq) GetWidth() int32 {
	if x != nil {
		return x.Width
	}
	return 0
}

func (x *CmdReq) GetPort() string {
	if x != nil {
		return x.Port
	}
	return ""
}

func (x *CmdReq) GetDaemon() bool {
	if x != nil {
		return x.Daemon
	}
	return false
}

type CommonReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Version string `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
}

func (x *CommonReq) Reset() {
	*x = CommonReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stream_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CommonReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CommonReq) ProtoMessage() {}

func (x *CommonReq) ProtoReflect() protoreflect.Message {
	mi := &file_stream_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CommonReq.ProtoReflect.Descriptor instead.
func (*CommonReq) Descriptor() ([]byte, []int) {
	return file_stream_proto_rawDescGZIP(), []int{3}
}

func (x *CommonReq) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

type CommonResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ok bool `protobuf:"varint,1,opt,name=ok,proto3" json:"ok,omitempty"`
}

func (x *CommonResp) Reset() {
	*x = CommonResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stream_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CommonResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CommonResp) ProtoMessage() {}

func (x *CommonResp) ProtoReflect() protoreflect.Message {
	mi := &file_stream_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CommonResp.ProtoReflect.Descriptor instead.
func (*CommonResp) Descriptor() ([]byte, []int) {
	return file_stream_proto_rawDescGZIP(), []int{4}
}

func (x *CommonResp) GetOk() bool {
	if x != nil {
		return x.Ok
	}
	return false
}

var File_stream_proto protoreflect.FileDescriptor

var file_stream_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02,
	0x70, 0x62, 0x22, 0x92, 0x02, 0x0a, 0x0c, 0x50, 0x75, 0x74, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d,
	0x52, 0x65, 0x71, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x64, 0x35, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x64, 0x35, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x6f, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6c, 0x6f, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x12, 0x0e, 0x0a, 0x02, 0x73, 0x6e, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x73, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x6f, 0x72,
	0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x1a, 0x0a,
	0x08, 0x6e, 0x6f, 0x64, 0x65, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x6e, 0x6f, 0x64, 0x65, 0x6c, 0x69, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x77, 0x69, 0x64,
	0x74, 0x68, 0x18, 0x08, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x77, 0x69, 0x64, 0x74, 0x68, 0x12,
	0x10, 0x0a, 0x03, 0x75, 0x69, 0x64, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x03, 0x75, 0x69,
	0x64, 0x12, 0x10, 0x0a, 0x03, 0x67, 0x69, 0x64, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x03,
	0x67, 0x69, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x66, 0x69, 0x6c, 0x65, 0x6d, 0x6f, 0x64, 0x18, 0x0b,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x66, 0x69, 0x6c, 0x65, 0x6d, 0x6f, 0x64, 0x12, 0x18, 0x0a,
	0x07, 0x6d, 0x6f, 0x64, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07,
	0x6d, 0x6f, 0x64, 0x74, 0x69, 0x6d, 0x65, 0x22, 0x49, 0x0a, 0x05, 0x52, 0x65, 0x70, 0x6c, 0x79,
	0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04,
	0x70, 0x61, 0x73, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x6e, 0x6f, 0x64, 0x65, 0x6c, 0x69, 0x73, 0x74,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6e, 0x6f, 0x64, 0x65, 0x6c, 0x69, 0x73, 0x74,
	0x12, 0x10, 0x0a, 0x03, 0x6d, 0x73, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d,
	0x73, 0x67, 0x22, 0x78, 0x0a, 0x06, 0x43, 0x6d, 0x64, 0x52, 0x65, 0x71, 0x12, 0x10, 0x0a, 0x03,
	0x63, 0x6d, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x63, 0x6d, 0x64, 0x12, 0x1a,
	0x0a, 0x08, 0x6e, 0x6f, 0x64, 0x65, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x6e, 0x6f, 0x64, 0x65, 0x6c, 0x69, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x77, 0x69,
	0x64, 0x74, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x77, 0x69, 0x64, 0x74, 0x68,
	0x12, 0x12, 0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x70, 0x6f, 0x72, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x22, 0x25, 0x0a, 0x09,
	0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73,
	0x69, 0x6f, 0x6e, 0x22, 0x1c, 0x0a, 0x0a, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x52, 0x65, 0x73,
	0x70, 0x12, 0x0e, 0x0a, 0x02, 0x6f, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x02, 0x6f,
	0x6b, 0x32, 0x84, 0x01, 0x0a, 0x0a, 0x52, 0x70, 0x63, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x12, 0x2c, 0x0a, 0x09, 0x50, 0x75, 0x74, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x12, 0x10, 0x2e,
	0x70, 0x62, 0x2e, 0x50, 0x75, 0x74, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x71, 0x1a,
	0x09, 0x2e, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x28, 0x01, 0x30, 0x01, 0x12, 0x21,
	0x0a, 0x06, 0x52, 0x75, 0x6e, 0x43, 0x6d, 0x64, 0x12, 0x0a, 0x2e, 0x70, 0x62, 0x2e, 0x43, 0x6d,
	0x64, 0x52, 0x65, 0x71, 0x1a, 0x09, 0x2e, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x30,
	0x01, 0x12, 0x25, 0x0a, 0x04, 0x50, 0x69, 0x6e, 0x67, 0x12, 0x0d, 0x2e, 0x70, 0x62, 0x2e, 0x43,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x1a, 0x0e, 0x2e, 0x70, 0x62, 0x2e, 0x43, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x42, 0x05, 0x5a, 0x03, 0x70, 0x62, 0x2f, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_stream_proto_rawDescOnce sync.Once
	file_stream_proto_rawDescData = file_stream_proto_rawDesc
)

func file_stream_proto_rawDescGZIP() []byte {
	file_stream_proto_rawDescOnce.Do(func() {
		file_stream_proto_rawDescData = protoimpl.X.CompressGZIP(file_stream_proto_rawDescData)
	})
	return file_stream_proto_rawDescData
}

var file_stream_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_stream_proto_goTypes = []interface{}{
	(*PutStreamReq)(nil), // 0: pb.PutStreamReq
	(*Reply)(nil),        // 1: pb.Reply
	(*CmdReq)(nil),       // 2: pb.CmdReq
	(*CommonReq)(nil),    // 3: pb.CommonReq
	(*CommonResp)(nil),   // 4: pb.CommonResp
}
var file_stream_proto_depIdxs = []int32{
	0, // 0: pb.RpcService.PutStream:input_type -> pb.PutStreamReq
	2, // 1: pb.RpcService.RunCmd:input_type -> pb.CmdReq
	3, // 2: pb.RpcService.Ping:input_type -> pb.CommonReq
	1, // 3: pb.RpcService.PutStream:output_type -> pb.Reply
	1, // 4: pb.RpcService.RunCmd:output_type -> pb.Reply
	4, // 5: pb.RpcService.Ping:output_type -> pb.CommonResp
	3, // [3:6] is the sub-list for method output_type
	0, // [0:3] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_stream_proto_init() }
func file_stream_proto_init() {
	if File_stream_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_stream_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PutStreamReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_stream_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Reply); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_stream_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CmdReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_stream_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CommonReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_stream_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CommonResp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_stream_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_stream_proto_goTypes,
		DependencyIndexes: file_stream_proto_depIdxs,
		MessageInfos:      file_stream_proto_msgTypes,
	}.Build()
	File_stream_proto = out.File
	file_stream_proto_rawDesc = nil
	file_stream_proto_goTypes = nil
	file_stream_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// RpcServiceClient is the client API for RpcService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
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
	stream, err := c.cc.NewStream(ctx, &_RpcService_serviceDesc.Streams[0], "/pb.RpcService/PutStream", opts...)
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
	stream, err := c.cc.NewStream(ctx, &_RpcService_serviceDesc.Streams[1], "/pb.RpcService/RunCmd", opts...)
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
type RpcServiceServer interface {
	PutStream(RpcService_PutStreamServer) error
	RunCmd(*CmdReq, RpcService_RunCmdServer) error
	Ping(context.Context, *CommonReq) (*CommonResp, error)
}

// UnimplementedRpcServiceServer can be embedded to have forward compatible implementations.
type UnimplementedRpcServiceServer struct {
}

func (*UnimplementedRpcServiceServer) PutStream(RpcService_PutStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method PutStream not implemented")
}
func (*UnimplementedRpcServiceServer) RunCmd(*CmdReq, RpcService_RunCmdServer) error {
	return status.Errorf(codes.Unimplemented, "method RunCmd not implemented")
}
func (*UnimplementedRpcServiceServer) Ping(context.Context, *CommonReq) (*CommonResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}

func RegisterRpcServiceServer(s *grpc.Server, srv RpcServiceServer) {
	s.RegisterService(&_RpcService_serviceDesc, srv)
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

var _RpcService_serviceDesc = grpc.ServiceDesc{
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
	Metadata: "stream.proto",
}
