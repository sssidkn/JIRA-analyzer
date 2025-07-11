// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: connector.proto

package connectorApi

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	JiraConnector_UpdateProject_FullMethodName = "/api.JiraConnector/UpdateProject"
	JiraConnector_GetProjects_FullMethodName   = "/api.JiraConnector/GetProjects"
)

// JiraConnectorClient is the client API for JiraConnector service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type JiraConnectorClient interface {
	UpdateProject(ctx context.Context, in *UpdateProjectRequest, opts ...grpc.CallOption) (*UpdateProjectResponse, error)
	GetProjects(ctx context.Context, in *GetProjectsRequest, opts ...grpc.CallOption) (*GetProjectsResponse, error)
}

type jiraConnectorClient struct {
	cc grpc.ClientConnInterface
}

func NewJiraConnectorClient(cc grpc.ClientConnInterface) JiraConnectorClient {
	return &jiraConnectorClient{cc}
}

func (c *jiraConnectorClient) UpdateProject(ctx context.Context, in *UpdateProjectRequest, opts ...grpc.CallOption) (*UpdateProjectResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateProjectResponse)
	err := c.cc.Invoke(ctx, JiraConnector_UpdateProject_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *jiraConnectorClient) GetProjects(ctx context.Context, in *GetProjectsRequest, opts ...grpc.CallOption) (*GetProjectsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetProjectsResponse)
	err := c.cc.Invoke(ctx, JiraConnector_GetProjects_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// JiraConnectorServer is the server API for JiraConnector service.
// All implementations must embed UnimplementedJiraConnectorServer
// for forward compatibility.
type JiraConnectorServer interface {
	UpdateProject(context.Context, *UpdateProjectRequest) (*UpdateProjectResponse, error)
	GetProjects(context.Context, *GetProjectsRequest) (*GetProjectsResponse, error)
	mustEmbedUnimplementedJiraConnectorServer()
}

// UnimplementedJiraConnectorServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedJiraConnectorServer struct{}

func (UnimplementedJiraConnectorServer) UpdateProject(context.Context, *UpdateProjectRequest) (*UpdateProjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateProject not implemented")
}
func (UnimplementedJiraConnectorServer) GetProjects(context.Context, *GetProjectsRequest) (*GetProjectsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetProjects not implemented")
}
func (UnimplementedJiraConnectorServer) mustEmbedUnimplementedJiraConnectorServer() {}
func (UnimplementedJiraConnectorServer) testEmbeddedByValue()                       {}

// UnsafeJiraConnectorServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to JiraConnectorServer will
// result in compilation errors.
type UnsafeJiraConnectorServer interface {
	mustEmbedUnimplementedJiraConnectorServer()
}

func RegisterJiraConnectorServer(s grpc.ServiceRegistrar, srv JiraConnectorServer) {
	// If the following call pancis, it indicates UnimplementedJiraConnectorServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&JiraConnector_ServiceDesc, srv)
}

func _JiraConnector_UpdateProject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateProjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JiraConnectorServer).UpdateProject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: JiraConnector_UpdateProject_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JiraConnectorServer).UpdateProject(ctx, req.(*UpdateProjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _JiraConnector_GetProjects_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetProjectsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JiraConnectorServer).GetProjects(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: JiraConnector_GetProjects_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JiraConnectorServer).GetProjects(ctx, req.(*GetProjectsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// JiraConnector_ServiceDesc is the grpc.ServiceDesc for JiraConnector service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var JiraConnector_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.JiraConnector",
	HandlerType: (*JiraConnectorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "UpdateProject",
			Handler:    _JiraConnector_UpdateProject_Handler,
		},
		{
			MethodName: "GetProjects",
			Handler:    _JiraConnector_GetProjects_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "connector.proto",
}
