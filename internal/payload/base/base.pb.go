// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v5.29.0--rc2
// source: base.proto

package base

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	auth "willofdaedalus/superluminal/internal/payload/auth"
	common "willofdaedalus/superluminal/internal/payload/common"
	error1 "willofdaedalus/superluminal/internal/payload/error"
	heartbeat "willofdaedalus/superluminal/internal/payload/heartbeat"
	info "willofdaedalus/superluminal/internal/payload/info"
	term "willofdaedalus/superluminal/internal/payload/term"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Payload struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Version   int32         `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	Header    common.Header `protobuf:"varint,2,opt,name=header,proto3,enum=Header" json:"header,omitempty"`
	Timestamp uint64        `protobuf:"varint,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	// Types that are assignable to Content:
	//
	//	*Payload_TermContent
	//	*Payload_Auth
	//	*Payload_Heartbeat
	//	*Payload_Error
	//	*Payload_Info
	Content isPayload_Content `protobuf_oneof:"content"`
}

func (x *Payload) Reset() {
	*x = Payload{}
	mi := &file_base_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Payload) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Payload) ProtoMessage() {}

func (x *Payload) ProtoReflect() protoreflect.Message {
	mi := &file_base_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Payload.ProtoReflect.Descriptor instead.
func (*Payload) Descriptor() ([]byte, []int) {
	return file_base_proto_rawDescGZIP(), []int{0}
}

func (x *Payload) GetVersion() int32 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *Payload) GetHeader() common.Header {
	if x != nil {
		return x.Header
	}
	return common.Header(0)
}

func (x *Payload) GetTimestamp() uint64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (m *Payload) GetContent() isPayload_Content {
	if m != nil {
		return m.Content
	}
	return nil
}

func (x *Payload) GetTermContent() *term.TerminalContent {
	if x, ok := x.GetContent().(*Payload_TermContent); ok {
		return x.TermContent
	}
	return nil
}

func (x *Payload) GetAuth() *auth.Authentication {
	if x, ok := x.GetContent().(*Payload_Auth); ok {
		return x.Auth
	}
	return nil
}

func (x *Payload) GetHeartbeat() *heartbeat.Heartbeat {
	if x, ok := x.GetContent().(*Payload_Heartbeat); ok {
		return x.Heartbeat
	}
	return nil
}

func (x *Payload) GetError() *error1.ErrorMessage {
	if x, ok := x.GetContent().(*Payload_Error); ok {
		return x.Error
	}
	return nil
}

func (x *Payload) GetInfo() *info.Info {
	if x, ok := x.GetContent().(*Payload_Info); ok {
		return x.Info
	}
	return nil
}

type isPayload_Content interface {
	isPayload_Content()
}

type Payload_TermContent struct {
	TermContent *term.TerminalContent `protobuf:"bytes,4,opt,name=term_content,json=termContent,proto3,oneof"`
}

type Payload_Auth struct {
	Auth *auth.Authentication `protobuf:"bytes,5,opt,name=auth,proto3,oneof"`
}

type Payload_Heartbeat struct {
	Heartbeat *heartbeat.Heartbeat `protobuf:"bytes,6,opt,name=heartbeat,proto3,oneof"`
}

type Payload_Error struct {
	Error *error1.ErrorMessage `protobuf:"bytes,7,opt,name=error,proto3,oneof"`
}

type Payload_Info struct {
	Info *info.Info `protobuf:"bytes,8,opt,name=info,proto3,oneof"`
}

func (*Payload_TermContent) isPayload_Content() {}

func (*Payload_Auth) isPayload_Content() {}

func (*Payload_Heartbeat) isPayload_Content() {}

func (*Payload_Error) isPayload_Content() {}

func (*Payload_Info) isPayload_Content() {}

var File_base_proto protoreflect.FileDescriptor

var file_base_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0a, 0x61, 0x75,
	0x74, 0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0c, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0b, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x0f, 0x68, 0x65, 0x61, 0x72, 0x74, 0x62, 0x65, 0x61, 0x74, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x12, 0x74, 0x65, 0x72, 0x6d, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0a, 0x69, 0x6e, 0x66, 0x6f, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0xbb, 0x02, 0x0a, 0x07, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64,
	0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x1f, 0x0a, 0x06, 0x68, 0x65,
	0x61, 0x64, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x07, 0x2e, 0x48, 0x65, 0x61,
	0x64, 0x65, 0x72, 0x52, 0x06, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x1c, 0x0a, 0x09, 0x74,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x09,
	0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x35, 0x0a, 0x0c, 0x74, 0x65, 0x72,
	0x6d, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x10, 0x2e, 0x54, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x61, 0x6c, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x48, 0x00, 0x52, 0x0b, 0x74, 0x65, 0x72, 0x6d, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74,
	0x12, 0x25, 0x0a, 0x04, 0x61, 0x75, 0x74, 0x68, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f,
	0x2e, 0x41, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x48,
	0x00, 0x52, 0x04, 0x61, 0x75, 0x74, 0x68, 0x12, 0x2a, 0x0a, 0x09, 0x68, 0x65, 0x61, 0x72, 0x74,
	0x62, 0x65, 0x61, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x48, 0x65, 0x61,
	0x72, 0x74, 0x62, 0x65, 0x61, 0x74, 0x48, 0x00, 0x52, 0x09, 0x68, 0x65, 0x61, 0x72, 0x74, 0x62,
	0x65, 0x61, 0x74, 0x12, 0x25, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x07, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x48, 0x00, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x1b, 0x0a, 0x04, 0x69, 0x6e,
	0x66, 0x6f, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x05, 0x2e, 0x49, 0x6e, 0x66, 0x6f, 0x48,
	0x00, 0x52, 0x04, 0x69, 0x6e, 0x66, 0x6f, 0x42, 0x09, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x42, 0x33, 0x5a, 0x31, 0x77, 0x69, 0x6c, 0x6c, 0x6f, 0x66, 0x64, 0x61, 0x65, 0x64,
	0x61, 0x6c, 0x75, 0x73, 0x2f, 0x73, 0x75, 0x70, 0x65, 0x72, 0x6c, 0x75, 0x6d, 0x69, 0x6e, 0x61,
	0x6c, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x61, 0x79, 0x6c, 0x6f,
	0x61, 0x64, 0x2f, 0x62, 0x61, 0x73, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_base_proto_rawDescOnce sync.Once
	file_base_proto_rawDescData = file_base_proto_rawDesc
)

func file_base_proto_rawDescGZIP() []byte {
	file_base_proto_rawDescOnce.Do(func() {
		file_base_proto_rawDescData = protoimpl.X.CompressGZIP(file_base_proto_rawDescData)
	})
	return file_base_proto_rawDescData
}

var file_base_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_base_proto_goTypes = []any{
	(*Payload)(nil),              // 0: Payload
	(common.Header)(0),           // 1: Header
	(*term.TerminalContent)(nil), // 2: TerminalContent
	(*auth.Authentication)(nil),  // 3: Authentication
	(*heartbeat.Heartbeat)(nil),  // 4: Heartbeat
	(*error1.ErrorMessage)(nil),  // 5: ErrorMessage
	(*info.Info)(nil),            // 6: Info
}
var file_base_proto_depIdxs = []int32{
	1, // 0: Payload.header:type_name -> Header
	2, // 1: Payload.term_content:type_name -> TerminalContent
	3, // 2: Payload.auth:type_name -> Authentication
	4, // 3: Payload.heartbeat:type_name -> Heartbeat
	5, // 4: Payload.error:type_name -> ErrorMessage
	6, // 5: Payload.info:type_name -> Info
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_base_proto_init() }
func file_base_proto_init() {
	if File_base_proto != nil {
		return
	}
	file_base_proto_msgTypes[0].OneofWrappers = []any{
		(*Payload_TermContent)(nil),
		(*Payload_Auth)(nil),
		(*Payload_Heartbeat)(nil),
		(*Payload_Error)(nil),
		(*Payload_Info)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_base_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_base_proto_goTypes,
		DependencyIndexes: file_base_proto_depIdxs,
		MessageInfos:      file_base_proto_msgTypes,
	}.Build()
	File_base_proto = out.File
	file_base_proto_rawDesc = nil
	file_base_proto_goTypes = nil
	file_base_proto_depIdxs = nil
}