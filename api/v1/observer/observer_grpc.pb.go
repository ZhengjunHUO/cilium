// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Hubble

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.24.0
// source: observer/observer.proto

package observer

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
	Observer_GetFlows_FullMethodName       = "/observer.Observer/GetFlows"
	Observer_GetAgentEvents_FullMethodName = "/observer.Observer/GetAgentEvents"
	Observer_GetDebugEvents_FullMethodName = "/observer.Observer/GetDebugEvents"
	Observer_GetNodes_FullMethodName       = "/observer.Observer/GetNodes"
	Observer_GetNamespaces_FullMethodName  = "/observer.Observer/GetNamespaces"
	Observer_ServerStatus_FullMethodName   = "/observer.Observer/ServerStatus"
)

// ObserverClient is the client API for Observer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ObserverClient interface {
	// GetFlows returning structured data, meant to eventually obsolete GetLastNFlows.
	GetFlows(ctx context.Context, in *GetFlowsRequest, opts ...grpc.CallOption) (Observer_GetFlowsClient, error)
	// GetAgentEvents returns Cilium agent events.
	GetAgentEvents(ctx context.Context, in *GetAgentEventsRequest, opts ...grpc.CallOption) (Observer_GetAgentEventsClient, error)
	// GetDebugEvents returns Cilium datapath debug events.
	GetDebugEvents(ctx context.Context, in *GetDebugEventsRequest, opts ...grpc.CallOption) (Observer_GetDebugEventsClient, error)
	// GetNodes returns information about nodes in a cluster.
	GetNodes(ctx context.Context, in *GetNodesRequest, opts ...grpc.CallOption) (*GetNodesResponse, error)
	// GetNamespaces returns information about namespaces in a cluster.
	// The namespaces returned are namespaces which have had network flows in
	// the last hour. The namespaces are returned sorted by cluster name and
	// namespace in ascending order.
	GetNamespaces(ctx context.Context, in *GetNamespacesRequest, opts ...grpc.CallOption) (*GetNamespacesResponse, error)
	// ServerStatus returns some details about the running hubble server.
	ServerStatus(ctx context.Context, in *ServerStatusRequest, opts ...grpc.CallOption) (*ServerStatusResponse, error)
}

type observerClient struct {
	cc grpc.ClientConnInterface
}

func NewObserverClient(cc grpc.ClientConnInterface) ObserverClient {
	return &observerClient{cc}
}

func (c *observerClient) GetFlows(ctx context.Context, in *GetFlowsRequest, opts ...grpc.CallOption) (Observer_GetFlowsClient, error) {
	stream, err := c.cc.NewStream(ctx, &Observer_ServiceDesc.Streams[0], Observer_GetFlows_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &observerGetFlowsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Observer_GetFlowsClient interface {
	Recv() (*GetFlowsResponse, error)
	grpc.ClientStream
}

type observerGetFlowsClient struct {
	grpc.ClientStream
}

func (x *observerGetFlowsClient) Recv() (*GetFlowsResponse, error) {
	m := new(GetFlowsResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *observerClient) GetAgentEvents(ctx context.Context, in *GetAgentEventsRequest, opts ...grpc.CallOption) (Observer_GetAgentEventsClient, error) {
	stream, err := c.cc.NewStream(ctx, &Observer_ServiceDesc.Streams[1], Observer_GetAgentEvents_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &observerGetAgentEventsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Observer_GetAgentEventsClient interface {
	Recv() (*GetAgentEventsResponse, error)
	grpc.ClientStream
}

type observerGetAgentEventsClient struct {
	grpc.ClientStream
}

func (x *observerGetAgentEventsClient) Recv() (*GetAgentEventsResponse, error) {
	m := new(GetAgentEventsResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *observerClient) GetDebugEvents(ctx context.Context, in *GetDebugEventsRequest, opts ...grpc.CallOption) (Observer_GetDebugEventsClient, error) {
	stream, err := c.cc.NewStream(ctx, &Observer_ServiceDesc.Streams[2], Observer_GetDebugEvents_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &observerGetDebugEventsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Observer_GetDebugEventsClient interface {
	Recv() (*GetDebugEventsResponse, error)
	grpc.ClientStream
}

type observerGetDebugEventsClient struct {
	grpc.ClientStream
}

func (x *observerGetDebugEventsClient) Recv() (*GetDebugEventsResponse, error) {
	m := new(GetDebugEventsResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *observerClient) GetNodes(ctx context.Context, in *GetNodesRequest, opts ...grpc.CallOption) (*GetNodesResponse, error) {
	out := new(GetNodesResponse)
	err := c.cc.Invoke(ctx, Observer_GetNodes_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *observerClient) GetNamespaces(ctx context.Context, in *GetNamespacesRequest, opts ...grpc.CallOption) (*GetNamespacesResponse, error) {
	out := new(GetNamespacesResponse)
	err := c.cc.Invoke(ctx, Observer_GetNamespaces_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *observerClient) ServerStatus(ctx context.Context, in *ServerStatusRequest, opts ...grpc.CallOption) (*ServerStatusResponse, error) {
	out := new(ServerStatusResponse)
	err := c.cc.Invoke(ctx, Observer_ServerStatus_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ObserverServer is the server API for Observer service.
// All implementations should embed UnimplementedObserverServer
// for forward compatibility
type ObserverServer interface {
	// GetFlows returning structured data, meant to eventually obsolete GetLastNFlows.
	GetFlows(*GetFlowsRequest, Observer_GetFlowsServer) error
	// GetAgentEvents returns Cilium agent events.
	GetAgentEvents(*GetAgentEventsRequest, Observer_GetAgentEventsServer) error
	// GetDebugEvents returns Cilium datapath debug events.
	GetDebugEvents(*GetDebugEventsRequest, Observer_GetDebugEventsServer) error
	// GetNodes returns information about nodes in a cluster.
	GetNodes(context.Context, *GetNodesRequest) (*GetNodesResponse, error)
	// GetNamespaces returns information about namespaces in a cluster.
	// The namespaces returned are namespaces which have had network flows in
	// the last hour. The namespaces are returned sorted by cluster name and
	// namespace in ascending order.
	GetNamespaces(context.Context, *GetNamespacesRequest) (*GetNamespacesResponse, error)
	// ServerStatus returns some details about the running hubble server.
	ServerStatus(context.Context, *ServerStatusRequest) (*ServerStatusResponse, error)
}

// UnimplementedObserverServer should be embedded to have forward compatible implementations.
type UnimplementedObserverServer struct {
}

func (UnimplementedObserverServer) GetFlows(*GetFlowsRequest, Observer_GetFlowsServer) error {
	return status.Errorf(codes.Unimplemented, "method GetFlows not implemented")
}
func (UnimplementedObserverServer) GetAgentEvents(*GetAgentEventsRequest, Observer_GetAgentEventsServer) error {
	return status.Errorf(codes.Unimplemented, "method GetAgentEvents not implemented")
}
func (UnimplementedObserverServer) GetDebugEvents(*GetDebugEventsRequest, Observer_GetDebugEventsServer) error {
	return status.Errorf(codes.Unimplemented, "method GetDebugEvents not implemented")
}
func (UnimplementedObserverServer) GetNodes(context.Context, *GetNodesRequest) (*GetNodesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNodes not implemented")
}
func (UnimplementedObserverServer) GetNamespaces(context.Context, *GetNamespacesRequest) (*GetNamespacesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNamespaces not implemented")
}
func (UnimplementedObserverServer) ServerStatus(context.Context, *ServerStatusRequest) (*ServerStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ServerStatus not implemented")
}

// UnsafeObserverServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ObserverServer will
// result in compilation errors.
type UnsafeObserverServer interface {
	mustEmbedUnimplementedObserverServer()
}

func RegisterObserverServer(s grpc.ServiceRegistrar, srv ObserverServer) {
	s.RegisterService(&Observer_ServiceDesc, srv)
}

func _Observer_GetFlows_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetFlowsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ObserverServer).GetFlows(m, &observerGetFlowsServer{stream})
}

type Observer_GetFlowsServer interface {
	Send(*GetFlowsResponse) error
	grpc.ServerStream
}

type observerGetFlowsServer struct {
	grpc.ServerStream
}

func (x *observerGetFlowsServer) Send(m *GetFlowsResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Observer_GetAgentEvents_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetAgentEventsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ObserverServer).GetAgentEvents(m, &observerGetAgentEventsServer{stream})
}

type Observer_GetAgentEventsServer interface {
	Send(*GetAgentEventsResponse) error
	grpc.ServerStream
}

type observerGetAgentEventsServer struct {
	grpc.ServerStream
}

func (x *observerGetAgentEventsServer) Send(m *GetAgentEventsResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Observer_GetDebugEvents_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetDebugEventsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ObserverServer).GetDebugEvents(m, &observerGetDebugEventsServer{stream})
}

type Observer_GetDebugEventsServer interface {
	Send(*GetDebugEventsResponse) error
	grpc.ServerStream
}

type observerGetDebugEventsServer struct {
	grpc.ServerStream
}

func (x *observerGetDebugEventsServer) Send(m *GetDebugEventsResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Observer_GetNodes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetNodesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObserverServer).GetNodes(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Observer_GetNodes_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObserverServer).GetNodes(ctx, req.(*GetNodesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Observer_GetNamespaces_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetNamespacesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObserverServer).GetNamespaces(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Observer_GetNamespaces_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObserverServer).GetNamespaces(ctx, req.(*GetNamespacesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Observer_ServerStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ServerStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObserverServer).ServerStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Observer_ServerStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObserverServer).ServerStatus(ctx, req.(*ServerStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Observer_ServiceDesc is the grpc.ServiceDesc for Observer service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Observer_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "observer.Observer",
	HandlerType: (*ObserverServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetNodes",
			Handler:    _Observer_GetNodes_Handler,
		},
		{
			MethodName: "GetNamespaces",
			Handler:    _Observer_GetNamespaces_Handler,
		},
		{
			MethodName: "ServerStatus",
			Handler:    _Observer_ServerStatus_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetFlows",
			Handler:       _Observer_GetFlows_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "GetAgentEvents",
			Handler:       _Observer_GetAgentEvents_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "GetDebugEvents",
			Handler:       _Observer_GetDebugEvents_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "observer/observer.proto",
}
