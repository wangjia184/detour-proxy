// Code generated by protoc-gen-go. DO NOT EDIT.
// source: dto.proto

/*
Package dto is a generated protocol buffer package.

It is generated from these files:
	dto.proto

It has these top-level messages:
	MessageHeader
	Payload
*/
package dto

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

type Type int32

const (
	Type_UNSPECIFIC                 Type = 0
	Type_TCP_CONNECT                Type = 1
	Type_TCP_CONNECTION_ESTABLISHED Type = 2
	Type_TCP_CONNECTION_FAILED      Type = 3
	Type_TCP_CONNECTION_CLOSED      Type = 4
	Type_INBOUND_DATA               Type = 5
	Type_OUTBOUND_DATA              Type = 6
)

var Type_name = map[int32]string{
	0: "UNSPECIFIC",
	1: "TCP_CONNECT",
	2: "TCP_CONNECTION_ESTABLISHED",
	3: "TCP_CONNECTION_FAILED",
	4: "TCP_CONNECTION_CLOSED",
	5: "INBOUND_DATA",
	6: "OUTBOUND_DATA",
}
var Type_value = map[string]int32{
	"UNSPECIFIC":                 0,
	"TCP_CONNECT":                1,
	"TCP_CONNECTION_ESTABLISHED": 2,
	"TCP_CONNECTION_FAILED":      3,
	"TCP_CONNECTION_CLOSED":      4,
	"INBOUND_DATA":               5,
	"OUTBOUND_DATA":              6,
}

func (x Type) String() string {
	return proto.EnumName(Type_name, int32(x))
}
func (Type) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type Mode int32

const (
	Mode_NONE  Mode = 0
	Mode_LZFSE Mode = 1
)

var Mode_name = map[int32]string{
	0: "NONE",
	1: "LZFSE",
}
var Mode_value = map[string]int32{
	"NONE":  0,
	"LZFSE": 1,
}

func (x Mode) String() string {
	return proto.EnumName(Mode_name, int32(x))
}
func (Mode) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type MessageHeader struct {
	Type         Type  `protobuf:"varint,1,opt,name=type,enum=dto.Type" json:"type,omitempty"`
	ConnectionID int64 `protobuf:"varint,2,opt,name=connectionID" json:"connectionID,omitempty"`
	Mode         Mode  `protobuf:"varint,3,opt,name=mode,enum=dto.Mode" json:"mode,omitempty"`
	Length       int32 `protobuf:"varint,4,opt,name=length" json:"length,omitempty"`
}

func (m *MessageHeader) Reset()                    { *m = MessageHeader{} }
func (m *MessageHeader) String() string            { return proto.CompactTextString(m) }
func (*MessageHeader) ProtoMessage()               {}
func (*MessageHeader) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *MessageHeader) GetType() Type {
	if m != nil {
		return m.Type
	}
	return Type_UNSPECIFIC
}

func (m *MessageHeader) GetConnectionID() int64 {
	if m != nil {
		return m.ConnectionID
	}
	return 0
}

func (m *MessageHeader) GetMode() Mode {
	if m != nil {
		return m.Mode
	}
	return Mode_NONE
}

func (m *MessageHeader) GetLength() int32 {
	if m != nil {
		return m.Length
	}
	return 0
}

type Payload struct {
	Address      string `protobuf:"bytes,1,opt,name=address" json:"address,omitempty"`
	Port         int32  `protobuf:"varint,2,opt,name=port" json:"port,omitempty"`
	Data         []byte `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
	ErrorMessage string `protobuf:"bytes,4,opt,name=errorMessage" json:"errorMessage,omitempty"`
}

func (m *Payload) Reset()                    { *m = Payload{} }
func (m *Payload) String() string            { return proto.CompactTextString(m) }
func (*Payload) ProtoMessage()               {}
func (*Payload) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Payload) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *Payload) GetPort() int32 {
	if m != nil {
		return m.Port
	}
	return 0
}

func (m *Payload) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *Payload) GetErrorMessage() string {
	if m != nil {
		return m.ErrorMessage
	}
	return ""
}

func init() {
	proto.RegisterType((*MessageHeader)(nil), "dto.MessageHeader")
	proto.RegisterType((*Payload)(nil), "dto.Payload")
	proto.RegisterEnum("dto.Type", Type_name, Type_value)
	proto.RegisterEnum("dto.Mode", Mode_name, Mode_value)
}

func init() { proto.RegisterFile("dto.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 359 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x91, 0xc1, 0x8e, 0xd3, 0x30,
	0x10, 0x86, 0xd7, 0x4d, 0xd2, 0x25, 0x43, 0x77, 0xd7, 0x8c, 0x04, 0x0a, 0x20, 0x50, 0xd5, 0x53,
	0xb5, 0x87, 0x1e, 0xe0, 0x09, 0xd2, 0xd8, 0xd5, 0x46, 0xca, 0x3a, 0x91, 0x93, 0x5e, 0xb8, 0x54,
	0x61, 0x6d, 0x2d, 0x48, 0x4b, 0x1c, 0x39, 0x46, 0x22, 0x8f, 0xc0, 0x6b, 0x70, 0xe4, 0x29, 0x51,
	0xac, 0x22, 0xb5, 0x88, 0xdb, 0xcc, 0x37, 0xf2, 0xfc, 0xff, 0xef, 0x81, 0x58, 0x39, 0xb3, 0xe9,
	0xad, 0x71, 0x06, 0x03, 0xe5, 0xcc, 0xea, 0x27, 0x81, 0xab, 0x7b, 0x3d, 0x0c, 0xed, 0xa3, 0xbe,
	0xd3, 0xad, 0xd2, 0x16, 0xdf, 0x41, 0xe8, 0xc6, 0x5e, 0x27, 0x64, 0x49, 0xd6, 0xd7, 0x1f, 0xe2,
	0xcd, 0xf4, 0xa0, 0x19, 0x7b, 0x2d, 0x3d, 0xc6, 0x15, 0x2c, 0x1e, 0x4c, 0xd7, 0xe9, 0x07, 0xf7,
	0xd5, 0x74, 0x39, 0x4b, 0x66, 0x4b, 0xb2, 0x0e, 0xe4, 0x19, 0x9b, 0x56, 0x7c, 0x33, 0x4a, 0x27,
	0xc1, 0xc9, 0x8a, 0x7b, 0xa3, 0xb4, 0xf4, 0x18, 0x5f, 0xc1, 0xfc, 0x49, 0x77, 0x8f, 0xee, 0x4b,
	0x12, 0x2e, 0xc9, 0x3a, 0x92, 0xc7, 0x6e, 0x65, 0xe0, 0xb2, 0x6a, 0xc7, 0x27, 0xd3, 0x2a, 0x4c,
	0xe0, 0xb2, 0x55, 0xca, 0xea, 0x61, 0xf0, 0x3e, 0x62, 0xf9, 0xb7, 0x45, 0x84, 0xb0, 0x37, 0xd6,
	0x79, 0xdd, 0x48, 0xfa, 0x7a, 0x62, 0xaa, 0x75, 0xad, 0xd7, 0x5b, 0x48, 0x5f, 0x4f, 0x3e, 0xb5,
	0xb5, 0xc6, 0x1e, 0xc3, 0x79, 0xa9, 0x58, 0x9e, 0xb1, 0xdb, 0x5f, 0x04, 0xc2, 0x29, 0x1a, 0x5e,
	0x03, 0xec, 0x45, 0x5d, 0xf1, 0x2c, 0xdf, 0xe5, 0x19, 0xbd, 0xc0, 0x1b, 0x78, 0xde, 0x64, 0xd5,
	0x21, 0x2b, 0x85, 0xe0, 0x59, 0x43, 0x09, 0xbe, 0x87, 0x37, 0x27, 0x20, 0x2f, 0xc5, 0x81, 0xd7,
	0x4d, 0xba, 0x2d, 0xf2, 0xfa, 0x8e, 0x33, 0x3a, 0xc3, 0xd7, 0xf0, 0xf2, 0x9f, 0xf9, 0x2e, 0xcd,
	0x0b, 0xce, 0x68, 0xf0, 0x9f, 0x51, 0x56, 0x94, 0x35, 0x67, 0x34, 0x44, 0x0a, 0x8b, 0x5c, 0x6c,
	0xcb, 0xbd, 0x60, 0x07, 0x96, 0x36, 0x29, 0x8d, 0xf0, 0x05, 0x5c, 0x95, 0xfb, 0xe6, 0x04, 0xcd,
	0x6f, 0xdf, 0x42, 0x38, 0xfd, 0x1d, 0x3e, 0x83, 0x50, 0x94, 0x82, 0xd3, 0x0b, 0x8c, 0x21, 0x2a,
	0x3e, 0xed, 0x6a, 0x4e, 0xc9, 0x16, 0x7f, 0xcf, 0x6e, 0x98, 0x76, 0xe6, 0xbb, 0xad, 0xac, 0xf9,
	0x31, 0x6e, 0x58, 0x53, 0x7e, 0x9e, 0xfb, 0xf3, 0x7e, 0xfc, 0x13, 0x00, 0x00, 0xff, 0xff, 0x5d,
	0x7c, 0xb3, 0x3a, 0xeb, 0x01, 0x00, 0x00,
}