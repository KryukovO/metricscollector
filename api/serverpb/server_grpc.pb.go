// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.15.8
// source: server.proto

package serverpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Storage_Update_FullMethodName     = "/server.Storage/Update"
	Storage_UpdateMany_FullMethodName = "/server.Storage/UpdateMany"
	Storage_Metric_FullMethodName     = "/server.Storage/Metric"
	Storage_AllMetrics_FullMethodName = "/server.Storage/AllMetrics"
)

// StorageClient is the client API for Storage service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type StorageClient interface {
	// Update выполняет обновление единственной метрики.
	Update(ctx context.Context, in *UpdateRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// UpdateMany выполняет обновления набора метрик.
	UpdateMany(ctx context.Context, in *UpdateManyRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// Metric возвращает описание метрики из хранилища.
	Metric(ctx context.Context, in *MetricRequest, opts ...grpc.CallOption) (*MetricResponse, error)
	// AllMetrics описание всех метрик из хранилища.
	AllMetrics(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*AllMetricsResponse, error)
}

type storageClient struct {
	cc grpc.ClientConnInterface
}

func NewStorageClient(cc grpc.ClientConnInterface) StorageClient {
	return &storageClient{cc}
}

func (c *storageClient) Update(ctx context.Context, in *UpdateRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Storage_Update_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storageClient) UpdateMany(ctx context.Context, in *UpdateManyRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Storage_UpdateMany_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storageClient) Metric(ctx context.Context, in *MetricRequest, opts ...grpc.CallOption) (*MetricResponse, error) {
	out := new(MetricResponse)
	err := c.cc.Invoke(ctx, Storage_Metric_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storageClient) AllMetrics(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*AllMetricsResponse, error) {
	out := new(AllMetricsResponse)
	err := c.cc.Invoke(ctx, Storage_AllMetrics_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// StorageServer is the server API for Storage service.
// All implementations must embed UnimplementedStorageServer
// for forward compatibility
type StorageServer interface {
	// Update выполняет обновление единственной метрики.
	Update(context.Context, *UpdateRequest) (*emptypb.Empty, error)
	// UpdateMany выполняет обновления набора метрик.
	UpdateMany(context.Context, *UpdateManyRequest) (*emptypb.Empty, error)
	// Metric возвращает описание метрики из хранилища.
	Metric(context.Context, *MetricRequest) (*MetricResponse, error)
	// AllMetrics описание всех метрик из хранилища.
	AllMetrics(context.Context, *emptypb.Empty) (*AllMetricsResponse, error)
	mustEmbedUnimplementedStorageServer()
}

// UnimplementedStorageServer must be embedded to have forward compatible implementations.
type UnimplementedStorageServer struct {
}

func (UnimplementedStorageServer) Update(context.Context, *UpdateRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Update not implemented")
}
func (UnimplementedStorageServer) UpdateMany(context.Context, *UpdateManyRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateMany not implemented")
}
func (UnimplementedStorageServer) Metric(context.Context, *MetricRequest) (*MetricResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Metric not implemented")
}
func (UnimplementedStorageServer) AllMetrics(context.Context, *emptypb.Empty) (*AllMetricsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AllMetrics not implemented")
}
func (UnimplementedStorageServer) mustEmbedUnimplementedStorageServer() {}

// UnsafeStorageServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to StorageServer will
// result in compilation errors.
type UnsafeStorageServer interface {
	mustEmbedUnimplementedStorageServer()
}

func RegisterStorageServer(s grpc.ServiceRegistrar, srv StorageServer) {
	s.RegisterService(&Storage_ServiceDesc, srv)
}

func _Storage_Update_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageServer).Update(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Storage_Update_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageServer).Update(ctx, req.(*UpdateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Storage_UpdateMany_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateManyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageServer).UpdateMany(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Storage_UpdateMany_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageServer).UpdateMany(ctx, req.(*UpdateManyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Storage_Metric_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MetricRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageServer).Metric(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Storage_Metric_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageServer).Metric(ctx, req.(*MetricRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Storage_AllMetrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageServer).AllMetrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Storage_AllMetrics_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageServer).AllMetrics(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// Storage_ServiceDesc is the grpc.ServiceDesc for Storage service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Storage_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "server.Storage",
	HandlerType: (*StorageServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Update",
			Handler:    _Storage_Update_Handler,
		},
		{
			MethodName: "UpdateMany",
			Handler:    _Storage_UpdateMany_Handler,
		},
		{
			MethodName: "Metric",
			Handler:    _Storage_Metric_Handler,
		},
		{
			MethodName: "AllMetrics",
			Handler:    _Storage_AllMetrics_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "server.proto",
}
