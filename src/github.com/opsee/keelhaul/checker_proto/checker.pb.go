// Code generated by protoc-gen-go.
// source: checker.proto
// DO NOT EDIT!

/*
Package checker is a generated protocol buffer package.

It is generated from these files:
	checker.proto

It has these top-level messages:
	Any
	Timestamp
	Target
	Check
	Header
	HttpCheck
	Metric
	HttpResponse
	CheckResourceResponse
	ResourceResponse
	CheckResourceRequest
	TestCheckRequest
	TestCheckResponse
	CheckResponse
	CheckResult
	DiscoveryEvent
*/
package checker_proto

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "go.pedge.io/google-protobuf"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type Any struct {
	TypeUrl string `protobuf:"bytes,1,opt,name=type_url" json:"type_url,omitempty"`
	Value   []byte `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *Any) Reset()         { *m = Any{} }
func (m *Any) String() string { return proto.CompactTextString(m) }
func (*Any) ProtoMessage()    {}

type Timestamp struct {
	Seconds int64 `protobuf:"varint,1,opt,name=seconds" json:"seconds,omitempty"`
	Nanos   int64 `protobuf:"varint,2,opt,name=nanos" json:"nanos,omitempty"`
}

func (m *Timestamp) Reset()         { *m = Timestamp{} }
func (m *Timestamp) String() string { return proto.CompactTextString(m) }
func (*Timestamp) ProtoMessage()    {}

type Target struct {
	Name    string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Type    string `protobuf:"bytes,2,opt,name=type" json:"type,omitempty"`
	Id      string `protobuf:"bytes,3,opt,name=id" json:"id,omitempty"`
	Address string `protobuf:"bytes,4,opt,name=address" json:"address,omitempty"`
}

func (m *Target) Reset()         { *m = Target{} }
func (m *Target) String() string { return proto.CompactTextString(m) }
func (*Target) ProtoMessage()    {}

type Check struct {
	Id        string     `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Interval  int32      `protobuf:"varint,2,opt,name=interval" json:"interval,omitempty"`
	Target    *Target    `protobuf:"bytes,3,opt,name=target" json:"target,omitempty"`
	LastRun   *Timestamp `protobuf:"bytes,4,opt,name=last_run" json:"last_run,omitempty"`
	CheckSpec *Any       `protobuf:"bytes,5,opt,name=check_spec" json:"check_spec,omitempty"`
	Name      string     `protobuf:"bytes,6,opt,name=name" json:"name,omitempty"`
}

func (m *Check) Reset()         { *m = Check{} }
func (m *Check) String() string { return proto.CompactTextString(m) }
func (*Check) ProtoMessage()    {}

func (m *Check) GetTarget() *Target {
	if m != nil {
		return m.Target
	}
	return nil
}

func (m *Check) GetLastRun() *Timestamp {
	if m != nil {
		return m.LastRun
	}
	return nil
}

func (m *Check) GetCheckSpec() *Any {
	if m != nil {
		return m.CheckSpec
	}
	return nil
}

type Header struct {
	Name   string   `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Values []string `protobuf:"bytes,2,rep,name=values" json:"values,omitempty"`
}

func (m *Header) Reset()         { *m = Header{} }
func (m *Header) String() string { return proto.CompactTextString(m) }
func (*Header) ProtoMessage()    {}

type HttpCheck struct {
	Name     string    `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Path     string    `protobuf:"bytes,2,opt,name=path" json:"path,omitempty"`
	Protocol string    `protobuf:"bytes,3,opt,name=protocol" json:"protocol,omitempty"`
	Port     int32     `protobuf:"varint,4,opt,name=port" json:"port,omitempty"`
	Verb     string    `protobuf:"bytes,5,opt,name=verb" json:"verb,omitempty"`
	Headers  []*Header `protobuf:"bytes,6,rep,name=headers" json:"headers,omitempty"`
	Body     string    `protobuf:"bytes,7,opt,name=body" json:"body,omitempty"`
}

func (m *HttpCheck) Reset()         { *m = HttpCheck{} }
func (m *HttpCheck) String() string { return proto.CompactTextString(m) }
func (*HttpCheck) ProtoMessage()    {}

func (m *HttpCheck) GetHeaders() []*Header {
	if m != nil {
		return m.Headers
	}
	return nil
}

type Metric struct {
	Name  string   `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Value float64  `protobuf:"fixed64,2,opt,name=value" json:"value,omitempty"`
	Tags  []string `protobuf:"bytes,3,rep,name=tags" json:"tags,omitempty"`
}

func (m *Metric) Reset()         { *m = Metric{} }
func (m *Metric) String() string { return proto.CompactTextString(m) }
func (*Metric) ProtoMessage()    {}

type HttpResponse struct {
	Code    int32     `protobuf:"varint,1,opt,name=code" json:"code,omitempty"`
	Body    string    `protobuf:"bytes,2,opt,name=body" json:"body,omitempty"`
	Headers []*Header `protobuf:"bytes,3,rep,name=headers" json:"headers,omitempty"`
	Metrics []*Metric `protobuf:"bytes,4,rep,name=metrics" json:"metrics,omitempty"`
	Host    string    `protobuf:"bytes,5,opt,name=host" json:"host,omitempty"`
}

func (m *HttpResponse) Reset()         { *m = HttpResponse{} }
func (m *HttpResponse) String() string { return proto.CompactTextString(m) }
func (*HttpResponse) ProtoMessage()    {}

func (m *HttpResponse) GetHeaders() []*Header {
	if m != nil {
		return m.Headers
	}
	return nil
}

func (m *HttpResponse) GetMetrics() []*Metric {
	if m != nil {
		return m.Metrics
	}
	return nil
}

type CheckResourceResponse struct {
	Id    string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Check *Check `protobuf:"bytes,2,opt,name=check" json:"check,omitempty"`
	Error string `protobuf:"bytes,3,opt,name=error" json:"error,omitempty"`
}

func (m *CheckResourceResponse) Reset()         { *m = CheckResourceResponse{} }
func (m *CheckResourceResponse) String() string { return proto.CompactTextString(m) }
func (*CheckResourceResponse) ProtoMessage()    {}

func (m *CheckResourceResponse) GetCheck() *Check {
	if m != nil {
		return m.Check
	}
	return nil
}

type ResourceResponse struct {
	Responses []*CheckResourceResponse `protobuf:"bytes,1,rep,name=responses" json:"responses,omitempty"`
}

func (m *ResourceResponse) Reset()         { *m = ResourceResponse{} }
func (m *ResourceResponse) String() string { return proto.CompactTextString(m) }
func (*ResourceResponse) ProtoMessage()    {}

func (m *ResourceResponse) GetResponses() []*CheckResourceResponse {
	if m != nil {
		return m.Responses
	}
	return nil
}

type CheckResourceRequest struct {
	Checks []*Check `protobuf:"bytes,1,rep,name=checks" json:"checks,omitempty"`
}

func (m *CheckResourceRequest) Reset()         { *m = CheckResourceRequest{} }
func (m *CheckResourceRequest) String() string { return proto.CompactTextString(m) }
func (*CheckResourceRequest) ProtoMessage()    {}

func (m *CheckResourceRequest) GetChecks() []*Check {
	if m != nil {
		return m.Checks
	}
	return nil
}

type TestCheckRequest struct {
	MaxHosts int32      `protobuf:"varint,1,opt,name=max_hosts" json:"max_hosts,omitempty"`
	Deadline *Timestamp `protobuf:"bytes,2,opt,name=deadline" json:"deadline,omitempty"`
	Check    *Check     `protobuf:"bytes,3,opt,name=check" json:"check,omitempty"`
}

func (m *TestCheckRequest) Reset()         { *m = TestCheckRequest{} }
func (m *TestCheckRequest) String() string { return proto.CompactTextString(m) }
func (*TestCheckRequest) ProtoMessage()    {}

func (m *TestCheckRequest) GetDeadline() *Timestamp {
	if m != nil {
		return m.Deadline
	}
	return nil
}

func (m *TestCheckRequest) GetCheck() *Check {
	if m != nil {
		return m.Check
	}
	return nil
}

type TestCheckResponse struct {
	Responses []*CheckResponse `protobuf:"bytes,1,rep,name=responses" json:"responses,omitempty"`
	Error     string           `protobuf:"bytes,2,opt,name=error" json:"error,omitempty"`
}

func (m *TestCheckResponse) Reset()         { *m = TestCheckResponse{} }
func (m *TestCheckResponse) String() string { return proto.CompactTextString(m) }
func (*TestCheckResponse) ProtoMessage()    {}

func (m *TestCheckResponse) GetResponses() []*CheckResponse {
	if m != nil {
		return m.Responses
	}
	return nil
}

type CheckResponse struct {
	Target   *Target `protobuf:"bytes,1,opt,name=target" json:"target,omitempty"`
	Response *Any    `protobuf:"bytes,2,opt,name=response" json:"response,omitempty"`
	Error    string  `protobuf:"bytes,3,opt,name=error" json:"error,omitempty"`
	Passing  bool    `protobuf:"varint,4,opt,name=passing" json:"passing,omitempty"`
}

func (m *CheckResponse) Reset()         { *m = CheckResponse{} }
func (m *CheckResponse) String() string { return proto.CompactTextString(m) }
func (*CheckResponse) ProtoMessage()    {}

func (m *CheckResponse) GetTarget() *Target {
	if m != nil {
		return m.Target
	}
	return nil
}

func (m *CheckResponse) GetResponse() *Any {
	if m != nil {
		return m.Response
	}
	return nil
}

type CheckResult struct {
	CheckId    string           `protobuf:"bytes,1,opt,name=check_id" json:"check_id,omitempty"`
	CustomerId string           `protobuf:"bytes,2,opt,name=customer_id" json:"customer_id,omitempty"`
	Timestamp  *Timestamp       `protobuf:"bytes,3,opt,name=timestamp" json:"timestamp,omitempty"`
	Passing    bool             `protobuf:"varint,4,opt,name=passing" json:"passing,omitempty"`
	Responses  []*CheckResponse `protobuf:"bytes,5,rep,name=responses" json:"responses,omitempty"`
	Target     *Target          `protobuf:"bytes,6,opt,name=target" json:"target,omitempty"`
	CheckName  string           `protobuf:"bytes,7,opt,name=check_name" json:"check_name,omitempty"`
}

func (m *CheckResult) Reset()         { *m = CheckResult{} }
func (m *CheckResult) String() string { return proto.CompactTextString(m) }
func (*CheckResult) ProtoMessage()    {}

func (m *CheckResult) GetTimestamp() *Timestamp {
	if m != nil {
		return m.Timestamp
	}
	return nil
}

func (m *CheckResult) GetResponses() []*CheckResponse {
	if m != nil {
		return m.Responses
	}
	return nil
}

func (m *CheckResult) GetTarget() *Target {
	if m != nil {
		return m.Target
	}
	return nil
}

type DiscoveryEvent struct {
	CustomerId string     `protobuf:"bytes,1,opt,name=customer_id" json:"customer_id,omitempty"`
	Timestamp  *Timestamp `protobuf:"bytes,2,opt,name=timestamp" json:"timestamp,omitempty"`
	Type       string     `protobuf:"bytes,3,opt,name=type" json:"type,omitempty"`
	Resource   string     `protobuf:"bytes,4,opt,name=resource" json:"resource,omitempty"`
}

func (m *DiscoveryEvent) Reset()         { *m = DiscoveryEvent{} }
func (m *DiscoveryEvent) String() string { return proto.CompactTextString(m) }
func (*DiscoveryEvent) ProtoMessage()    {}

func (m *DiscoveryEvent) GetTimestamp() *Timestamp {
	if m != nil {
		return m.Timestamp
	}
	return nil
}

var E_IsRequired = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.FieldOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         60000,
	Name:          "checker.is_required",
	Tag:           "varint,60000,opt,name=is_required",
}

func init() {
	proto.RegisterExtension(E_IsRequired)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// Client API for Checker service

type CheckerClient interface {
	TestCheck(ctx context.Context, in *TestCheckRequest, opts ...grpc.CallOption) (*TestCheckResponse, error)
	CreateCheck(ctx context.Context, in *CheckResourceRequest, opts ...grpc.CallOption) (*ResourceResponse, error)
	RetrieveCheck(ctx context.Context, in *CheckResourceRequest, opts ...grpc.CallOption) (*ResourceResponse, error)
	UpdateCheck(ctx context.Context, in *CheckResourceRequest, opts ...grpc.CallOption) (*ResourceResponse, error)
	DeleteCheck(ctx context.Context, in *CheckResourceRequest, opts ...grpc.CallOption) (*ResourceResponse, error)
}

type checkerClient struct {
	cc *grpc.ClientConn
}

func NewCheckerClient(cc *grpc.ClientConn) CheckerClient {
	return &checkerClient{cc}
}

func (c *checkerClient) TestCheck(ctx context.Context, in *TestCheckRequest, opts ...grpc.CallOption) (*TestCheckResponse, error) {
	out := new(TestCheckResponse)
	err := grpc.Invoke(ctx, "/checker.Checker/TestCheck", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *checkerClient) CreateCheck(ctx context.Context, in *CheckResourceRequest, opts ...grpc.CallOption) (*ResourceResponse, error) {
	out := new(ResourceResponse)
	err := grpc.Invoke(ctx, "/checker.Checker/CreateCheck", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *checkerClient) RetrieveCheck(ctx context.Context, in *CheckResourceRequest, opts ...grpc.CallOption) (*ResourceResponse, error) {
	out := new(ResourceResponse)
	err := grpc.Invoke(ctx, "/checker.Checker/RetrieveCheck", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *checkerClient) UpdateCheck(ctx context.Context, in *CheckResourceRequest, opts ...grpc.CallOption) (*ResourceResponse, error) {
	out := new(ResourceResponse)
	err := grpc.Invoke(ctx, "/checker.Checker/UpdateCheck", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *checkerClient) DeleteCheck(ctx context.Context, in *CheckResourceRequest, opts ...grpc.CallOption) (*ResourceResponse, error) {
	out := new(ResourceResponse)
	err := grpc.Invoke(ctx, "/checker.Checker/DeleteCheck", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Checker service

type CheckerServer interface {
	TestCheck(context.Context, *TestCheckRequest) (*TestCheckResponse, error)
	CreateCheck(context.Context, *CheckResourceRequest) (*ResourceResponse, error)
	RetrieveCheck(context.Context, *CheckResourceRequest) (*ResourceResponse, error)
	UpdateCheck(context.Context, *CheckResourceRequest) (*ResourceResponse, error)
	DeleteCheck(context.Context, *CheckResourceRequest) (*ResourceResponse, error)
}

func RegisterCheckerServer(s *grpc.Server, srv CheckerServer) {
	s.RegisterService(&_Checker_serviceDesc, srv)
}

func _Checker_TestCheck_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(TestCheckRequest)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(CheckerServer).TestCheck(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _Checker_CreateCheck_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(CheckResourceRequest)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(CheckerServer).CreateCheck(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _Checker_RetrieveCheck_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(CheckResourceRequest)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(CheckerServer).RetrieveCheck(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _Checker_UpdateCheck_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(CheckResourceRequest)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(CheckerServer).UpdateCheck(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _Checker_DeleteCheck_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(CheckResourceRequest)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(CheckerServer).DeleteCheck(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

var _Checker_serviceDesc = grpc.ServiceDesc{
	ServiceName: "checker.Checker",
	HandlerType: (*CheckerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "TestCheck",
			Handler:    _Checker_TestCheck_Handler,
		},
		{
			MethodName: "CreateCheck",
			Handler:    _Checker_CreateCheck_Handler,
		},
		{
			MethodName: "RetrieveCheck",
			Handler:    _Checker_RetrieveCheck_Handler,
		},
		{
			MethodName: "UpdateCheck",
			Handler:    _Checker_UpdateCheck_Handler,
		},
		{
			MethodName: "DeleteCheck",
			Handler:    _Checker_DeleteCheck_Handler,
		},
	},
	Streams: []grpc.StreamDesc{},
}
