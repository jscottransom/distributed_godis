// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.23.3
// source: api/godis.proto

package godis

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

const (
	GodisService_SetKey_FullMethodName   = "/godis.GodisService/SetKey"
	GodisService_GetKey_FullMethodName   = "/godis.GodisService/GetKey"
	GodisService_ListKeys_FullMethodName = "/godis.GodisService/ListKeys"
)

// GodisServiceClient is the client API for GodisService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GodisServiceClient interface {
	SetKey(ctx context.Context, in *SetRequest, opts ...grpc.CallOption) (*SetResponse, error)
	GetKey(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error)
	ListKeys(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error)
}

type godisServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewGodisServiceClient(cc grpc.ClientConnInterface) GodisServiceClient {
	return &godisServiceClient{cc}
}

func (c *godisServiceClient) SetKey(ctx context.Context, in *SetRequest, opts ...grpc.CallOption) (*SetResponse, error) {
	out := new(SetResponse)
	err := c.cc.Invoke(ctx, GodisService_SetKey_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *godisServiceClient) GetKey(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error) {
	out := new(GetResponse)
	err := c.cc.Invoke(ctx, GodisService_GetKey_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *godisServiceClient) ListKeys(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error) {
	out := new(ListResponse)
	err := c.cc.Invoke(ctx, GodisService_ListKeys_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GodisServiceServer is the server API for GodisService service.
// All implementations must embed UnimplementedGodisServiceServer
// for forward compatibility
type GodisServiceServer interface {
	SetKey(context.Context, *SetRequest) (*SetResponse, error)
	GetKey(context.Context, *GetRequest) (*GetResponse, error)
	ListKeys(context.Context, *ListRequest) (*ListResponse, error)
	mustEmbedUnimplementedGodisServiceServer()
}

// UnimplementedGodisServiceServer must be embedded to have forward compatible implementations.
type UnimplementedGodisServiceServer struct {
}

func (UnimplementedGodisServiceServer) SetKey(context.Context, *SetRequest) (*SetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetKey not implemented")
}
func (UnimplementedGodisServiceServer) GetKey(context.Context, *GetRequest) (*GetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetKey not implemented")
}
func (UnimplementedGodisServiceServer) ListKeys(context.Context, *ListRequest) (*ListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListKeys not implemented")
}
func (UnimplementedGodisServiceServer) mustEmbedUnimplementedGodisServiceServer() {}

// UnsafeGodisServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GodisServiceServer will
// result in compilation errors.
type UnsafeGodisServiceServer interface {
	mustEmbedUnimplementedGodisServiceServer()
}

func RegisterGodisServiceServer(s grpc.ServiceRegistrar, srv GodisServiceServer) {
	s.RegisterService(&GodisService_ServiceDesc, srv)
}

func _GodisService_SetKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GodisServiceServer).SetKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GodisService_SetKey_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GodisServiceServer).SetKey(ctx, req.(*SetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GodisService_GetKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GodisServiceServer).GetKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GodisService_GetKey_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GodisServiceServer).GetKey(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GodisService_ListKeys_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GodisServiceServer).ListKeys(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GodisService_ListKeys_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GodisServiceServer).ListKeys(ctx, req.(*ListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// GodisService_ServiceDesc is the grpc.ServiceDesc for GodisService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GodisService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "godis.GodisService",
	HandlerType: (*GodisServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetKey",
			Handler:    _GodisService_SetKey_Handler,
		},
		{
			MethodName: "GetKey",
			Handler:    _GodisService_GetKey_Handler,
		},
		{
			MethodName: "ListKeys",
			Handler:    _GodisService_ListKeys_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/godis.proto",
}
