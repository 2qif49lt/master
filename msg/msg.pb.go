// Code generated by protoc-gen-go.
// source: msg.proto
// DO NOT EDIT!

/*
Package msg is a generated protocol buffer package.

It is generated from these files:
	msg.proto

It has these top-level messages:
	LoginReq
	LoginRsp
	AliveReq
	AliveRsp
	LogoutReq
	LogoutRsp
	NewConnPushReq
	NewConnPushRsp
	NewConnHandShakeReq
	NewConnHandShakeRsp
*/
package msg

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type LoginReq struct {
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Id   string `protobuf:"bytes,2,opt,name=id" json:"id,omitempty"`
}

func (m *LoginReq) Reset()                    { *m = LoginReq{} }
func (m *LoginReq) String() string            { return proto.CompactTextString(m) }
func (*LoginReq) ProtoMessage()               {}
func (*LoginReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type LoginRsp struct {
	Rst int32  `protobuf:"varint,1,opt,name=rst" json:"rst,omitempty"`
	Id  string `protobuf:"bytes,2,opt,name=id" json:"id,omitempty"`
}

func (m *LoginRsp) Reset()                    { *m = LoginRsp{} }
func (m *LoginRsp) String() string            { return proto.CompactTextString(m) }
func (*LoginRsp) ProtoMessage()               {}
func (*LoginRsp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type AliveReq struct {
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
}

func (m *AliveReq) Reset()                    { *m = AliveReq{} }
func (m *AliveReq) String() string            { return proto.CompactTextString(m) }
func (*AliveReq) ProtoMessage()               {}
func (*AliveReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type AliveRsp struct {
	Rst int32 `protobuf:"varint,1,opt,name=rst" json:"rst,omitempty"`
}

func (m *AliveRsp) Reset()                    { *m = AliveRsp{} }
func (m *AliveRsp) String() string            { return proto.CompactTextString(m) }
func (*AliveRsp) ProtoMessage()               {}
func (*AliveRsp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

type LogoutReq struct {
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
}

func (m *LogoutReq) Reset()                    { *m = LogoutReq{} }
func (m *LogoutReq) String() string            { return proto.CompactTextString(m) }
func (*LogoutReq) ProtoMessage()               {}
func (*LogoutReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type LogoutRsp struct {
	Rst int32 `protobuf:"varint,1,opt,name=rst" json:"rst,omitempty"`
}

func (m *LogoutRsp) Reset()                    { *m = LogoutRsp{} }
func (m *LogoutRsp) String() string            { return proto.CompactTextString(m) }
func (*LogoutRsp) ProtoMessage()               {}
func (*LogoutRsp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

// udpsrv send to client
type NewConnPushReq struct {
	Connid int32 `protobuf:"varint,1,opt,name=connid" json:"connid,omitempty"`
}

func (m *NewConnPushReq) Reset()                    { *m = NewConnPushReq{} }
func (m *NewConnPushReq) String() string            { return proto.CompactTextString(m) }
func (*NewConnPushReq) ProtoMessage()               {}
func (*NewConnPushReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

type NewConnPushRsp struct {
	Rst    int32 `protobuf:"varint,1,opt,name=rst" json:"rst,omitempty"`
	Connid int32 `protobuf:"varint,2,opt,name=connid" json:"connid,omitempty"`
}

func (m *NewConnPushRsp) Reset()                    { *m = NewConnPushRsp{} }
func (m *NewConnPushRsp) String() string            { return proto.CompactTextString(m) }
func (*NewConnPushRsp) ProtoMessage()               {}
func (*NewConnPushRsp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

// tcp message send by client
type NewConnHandShakeReq struct {
	Connid int32 `protobuf:"varint,1,opt,name=connid" json:"connid,omitempty"`
}

func (m *NewConnHandShakeReq) Reset()                    { *m = NewConnHandShakeReq{} }
func (m *NewConnHandShakeReq) String() string            { return proto.CompactTextString(m) }
func (*NewConnHandShakeReq) ProtoMessage()               {}
func (*NewConnHandShakeReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

type NewConnHandShakeRsp struct {
	Rst    int32 `protobuf:"varint,1,opt,name=rst" json:"rst,omitempty"`
	Connid int32 `protobuf:"varint,2,opt,name=connid" json:"connid,omitempty"`
}

func (m *NewConnHandShakeRsp) Reset()                    { *m = NewConnHandShakeRsp{} }
func (m *NewConnHandShakeRsp) String() string            { return proto.CompactTextString(m) }
func (*NewConnHandShakeRsp) ProtoMessage()               {}
func (*NewConnHandShakeRsp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func init() {
	proto.RegisterType((*LoginReq)(nil), "LoginReq")
	proto.RegisterType((*LoginRsp)(nil), "LoginRsp")
	proto.RegisterType((*AliveReq)(nil), "AliveReq")
	proto.RegisterType((*AliveRsp)(nil), "AliveRsp")
	proto.RegisterType((*LogoutReq)(nil), "LogoutReq")
	proto.RegisterType((*LogoutRsp)(nil), "LogoutRsp")
	proto.RegisterType((*NewConnPushReq)(nil), "NewConnPushReq")
	proto.RegisterType((*NewConnPushRsp)(nil), "NewConnPushRsp")
	proto.RegisterType((*NewConnHandShakeReq)(nil), "NewConnHandShakeReq")
	proto.RegisterType((*NewConnHandShakeRsp)(nil), "NewConnHandShakeRsp")
}

func init() { proto.RegisterFile("msg.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 203 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0xe2, 0xcc, 0x2d, 0x4e, 0xd7,
	0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x57, 0xd2, 0xe3, 0xe2, 0xf0, 0xc9, 0x4f, 0xcf, 0xcc, 0x0b, 0x4a,
	0x2d, 0x14, 0x12, 0xe2, 0x62, 0xc9, 0x4b, 0xcc, 0x4d, 0x95, 0x60, 0x54, 0x60, 0xd4, 0xe0, 0x0c,
	0x02, 0xb3, 0x85, 0xf8, 0xb8, 0x98, 0x32, 0x53, 0x24, 0x98, 0xc0, 0x22, 0x4c, 0x99, 0x29, 0x4a,
	0x3a, 0x30, 0xf5, 0xc5, 0x05, 0x42, 0x02, 0x5c, 0xcc, 0x45, 0xc5, 0x25, 0x60, 0xe5, 0xac, 0x41,
	0x20, 0x26, 0x86, 0x6a, 0x29, 0x2e, 0x0e, 0xc7, 0x9c, 0xcc, 0xb2, 0x54, 0x90, 0xe9, 0x10, 0x39,
	0x46, 0xb8, 0x9c, 0x0c, 0x4c, 0x0e, 0x9b, 0x49, 0x4a, 0xd2, 0x5c, 0x9c, 0x3e, 0xf9, 0xe9, 0xf9,
	0xa5, 0x25, 0xd8, 0xb4, 0xca, 0xc2, 0x25, 0xb1, 0xea, 0xd5, 0xe0, 0xe2, 0xf3, 0x4b, 0x2d, 0x77,
	0xce, 0xcf, 0xcb, 0x0b, 0x28, 0x2d, 0xce, 0x00, 0x19, 0x20, 0xc6, 0xc5, 0x96, 0x9c, 0x9f, 0x97,
	0x07, 0x35, 0x84, 0x35, 0x08, 0xca, 0x53, 0xb2, 0x42, 0x55, 0x89, 0xd5, 0x4f, 0x08, 0xbd, 0x4c,
	0x28, 0x7a, 0x75, 0xb9, 0x84, 0xa1, 0x7a, 0x3d, 0x12, 0xf3, 0x52, 0x82, 0x33, 0x12, 0xb3, 0x53,
	0xf1, 0x59, 0x65, 0x8f, 0x45, 0x39, 0x29, 0xf6, 0x25, 0xb1, 0x81, 0x23, 0xcc, 0x18, 0x10, 0x00,
	0x00, 0xff, 0xff, 0x69, 0x94, 0xb9, 0xa0, 0xbd, 0x01, 0x00, 0x00,
}
