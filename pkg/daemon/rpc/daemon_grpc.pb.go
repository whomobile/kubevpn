// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.21.2
// source: daemon.proto

package rpc

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
	Daemon_Connect_FullMethodName    = "/rpc.Daemon/Connect"
	Daemon_Disconnect_FullMethodName = "/rpc.Daemon/Disconnect"
	Daemon_Proxy_FullMethodName      = "/rpc.Daemon/Proxy"
	Daemon_Leave_FullMethodName      = "/rpc.Daemon/Leave"
	Daemon_Clone_FullMethodName      = "/rpc.Daemon/Clone"
	Daemon_Remove_FullMethodName     = "/rpc.Daemon/Remove"
	Daemon_Logs_FullMethodName       = "/rpc.Daemon/Logs"
	Daemon_List_FullMethodName       = "/rpc.Daemon/List"
	Daemon_Upgrade_FullMethodName    = "/rpc.Daemon/Upgrade"
	Daemon_Status_FullMethodName     = "/rpc.Daemon/Status"
	Daemon_Quit_FullMethodName       = "/rpc.Daemon/Quit"
)

// DaemonClient is the client API for Daemon service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DaemonClient interface {
	Connect(ctx context.Context, in *ConnectRequest, opts ...grpc.CallOption) (Daemon_ConnectClient, error)
	Disconnect(ctx context.Context, in *DisconnectRequest, opts ...grpc.CallOption) (Daemon_DisconnectClient, error)
	Proxy(ctx context.Context, in *ConnectRequest, opts ...grpc.CallOption) (Daemon_ProxyClient, error)
	Leave(ctx context.Context, in *LeaveRequest, opts ...grpc.CallOption) (Daemon_LeaveClient, error)
	Clone(ctx context.Context, in *CloneRequest, opts ...grpc.CallOption) (Daemon_CloneClient, error)
	Remove(ctx context.Context, in *RemoveRequest, opts ...grpc.CallOption) (Daemon_RemoveClient, error)
	Logs(ctx context.Context, in *LogRequest, opts ...grpc.CallOption) (Daemon_LogsClient, error)
	List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error)
	Upgrade(ctx context.Context, in *UpgradeRequest, opts ...grpc.CallOption) (Daemon_UpgradeClient, error)
	Status(ctx context.Context, in *StatusRequest, opts ...grpc.CallOption) (*StatusResponse, error)
	Quit(ctx context.Context, in *QuitRequest, opts ...grpc.CallOption) (Daemon_QuitClient, error)
}

type daemonClient struct {
	cc grpc.ClientConnInterface
}

func NewDaemonClient(cc grpc.ClientConnInterface) DaemonClient {
	return &daemonClient{cc}
}

func (c *daemonClient) Connect(ctx context.Context, in *ConnectRequest, opts ...grpc.CallOption) (Daemon_ConnectClient, error) {
	stream, err := c.cc.NewStream(ctx, &Daemon_ServiceDesc.Streams[0], Daemon_Connect_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &daemonConnectClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Daemon_ConnectClient interface {
	Recv() (*ConnectResponse, error)
	grpc.ClientStream
}

type daemonConnectClient struct {
	grpc.ClientStream
}

func (x *daemonConnectClient) Recv() (*ConnectResponse, error) {
	m := new(ConnectResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *daemonClient) Disconnect(ctx context.Context, in *DisconnectRequest, opts ...grpc.CallOption) (Daemon_DisconnectClient, error) {
	stream, err := c.cc.NewStream(ctx, &Daemon_ServiceDesc.Streams[1], Daemon_Disconnect_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &daemonDisconnectClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Daemon_DisconnectClient interface {
	Recv() (*DisconnectResponse, error)
	grpc.ClientStream
}

type daemonDisconnectClient struct {
	grpc.ClientStream
}

func (x *daemonDisconnectClient) Recv() (*DisconnectResponse, error) {
	m := new(DisconnectResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *daemonClient) Proxy(ctx context.Context, in *ConnectRequest, opts ...grpc.CallOption) (Daemon_ProxyClient, error) {
	stream, err := c.cc.NewStream(ctx, &Daemon_ServiceDesc.Streams[2], Daemon_Proxy_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &daemonProxyClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Daemon_ProxyClient interface {
	Recv() (*ConnectResponse, error)
	grpc.ClientStream
}

type daemonProxyClient struct {
	grpc.ClientStream
}

func (x *daemonProxyClient) Recv() (*ConnectResponse, error) {
	m := new(ConnectResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *daemonClient) Leave(ctx context.Context, in *LeaveRequest, opts ...grpc.CallOption) (Daemon_LeaveClient, error) {
	stream, err := c.cc.NewStream(ctx, &Daemon_ServiceDesc.Streams[3], Daemon_Leave_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &daemonLeaveClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Daemon_LeaveClient interface {
	Recv() (*LeaveResponse, error)
	grpc.ClientStream
}

type daemonLeaveClient struct {
	grpc.ClientStream
}

func (x *daemonLeaveClient) Recv() (*LeaveResponse, error) {
	m := new(LeaveResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *daemonClient) Clone(ctx context.Context, in *CloneRequest, opts ...grpc.CallOption) (Daemon_CloneClient, error) {
	stream, err := c.cc.NewStream(ctx, &Daemon_ServiceDesc.Streams[4], Daemon_Clone_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &daemonCloneClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Daemon_CloneClient interface {
	Recv() (*CloneResponse, error)
	grpc.ClientStream
}

type daemonCloneClient struct {
	grpc.ClientStream
}

func (x *daemonCloneClient) Recv() (*CloneResponse, error) {
	m := new(CloneResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *daemonClient) Remove(ctx context.Context, in *RemoveRequest, opts ...grpc.CallOption) (Daemon_RemoveClient, error) {
	stream, err := c.cc.NewStream(ctx, &Daemon_ServiceDesc.Streams[5], Daemon_Remove_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &daemonRemoveClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Daemon_RemoveClient interface {
	Recv() (*RemoveResponse, error)
	grpc.ClientStream
}

type daemonRemoveClient struct {
	grpc.ClientStream
}

func (x *daemonRemoveClient) Recv() (*RemoveResponse, error) {
	m := new(RemoveResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *daemonClient) Logs(ctx context.Context, in *LogRequest, opts ...grpc.CallOption) (Daemon_LogsClient, error) {
	stream, err := c.cc.NewStream(ctx, &Daemon_ServiceDesc.Streams[6], Daemon_Logs_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &daemonLogsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Daemon_LogsClient interface {
	Recv() (*LogResponse, error)
	grpc.ClientStream
}

type daemonLogsClient struct {
	grpc.ClientStream
}

func (x *daemonLogsClient) Recv() (*LogResponse, error) {
	m := new(LogResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *daemonClient) List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error) {
	out := new(ListResponse)
	err := c.cc.Invoke(ctx, Daemon_List_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *daemonClient) Upgrade(ctx context.Context, in *UpgradeRequest, opts ...grpc.CallOption) (Daemon_UpgradeClient, error) {
	stream, err := c.cc.NewStream(ctx, &Daemon_ServiceDesc.Streams[7], Daemon_Upgrade_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &daemonUpgradeClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Daemon_UpgradeClient interface {
	Recv() (*UpgradeResponse, error)
	grpc.ClientStream
}

type daemonUpgradeClient struct {
	grpc.ClientStream
}

func (x *daemonUpgradeClient) Recv() (*UpgradeResponse, error) {
	m := new(UpgradeResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *daemonClient) Status(ctx context.Context, in *StatusRequest, opts ...grpc.CallOption) (*StatusResponse, error) {
	out := new(StatusResponse)
	err := c.cc.Invoke(ctx, Daemon_Status_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *daemonClient) Quit(ctx context.Context, in *QuitRequest, opts ...grpc.CallOption) (Daemon_QuitClient, error) {
	stream, err := c.cc.NewStream(ctx, &Daemon_ServiceDesc.Streams[8], Daemon_Quit_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &daemonQuitClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Daemon_QuitClient interface {
	Recv() (*QuitResponse, error)
	grpc.ClientStream
}

type daemonQuitClient struct {
	grpc.ClientStream
}

func (x *daemonQuitClient) Recv() (*QuitResponse, error) {
	m := new(QuitResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// DaemonServer is the server API for Daemon service.
// All implementations must embed UnimplementedDaemonServer
// for forward compatibility
type DaemonServer interface {
	Connect(*ConnectRequest, Daemon_ConnectServer) error
	Disconnect(*DisconnectRequest, Daemon_DisconnectServer) error
	Proxy(*ConnectRequest, Daemon_ProxyServer) error
	Leave(*LeaveRequest, Daemon_LeaveServer) error
	Clone(*CloneRequest, Daemon_CloneServer) error
	Remove(*RemoveRequest, Daemon_RemoveServer) error
	Logs(*LogRequest, Daemon_LogsServer) error
	List(context.Context, *ListRequest) (*ListResponse, error)
	Upgrade(*UpgradeRequest, Daemon_UpgradeServer) error
	Status(context.Context, *StatusRequest) (*StatusResponse, error)
	Quit(*QuitRequest, Daemon_QuitServer) error
	mustEmbedUnimplementedDaemonServer()
}

// UnimplementedDaemonServer must be embedded to have forward compatible implementations.
type UnimplementedDaemonServer struct {
}

func (UnimplementedDaemonServer) Connect(*ConnectRequest, Daemon_ConnectServer) error {
	return status.Errorf(codes.Unimplemented, "method Connect not implemented")
}
func (UnimplementedDaemonServer) Disconnect(*DisconnectRequest, Daemon_DisconnectServer) error {
	return status.Errorf(codes.Unimplemented, "method Disconnect not implemented")
}
func (UnimplementedDaemonServer) Proxy(*ConnectRequest, Daemon_ProxyServer) error {
	return status.Errorf(codes.Unimplemented, "method Proxy not implemented")
}
func (UnimplementedDaemonServer) Leave(*LeaveRequest, Daemon_LeaveServer) error {
	return status.Errorf(codes.Unimplemented, "method Leave not implemented")
}
func (UnimplementedDaemonServer) Clone(*CloneRequest, Daemon_CloneServer) error {
	return status.Errorf(codes.Unimplemented, "method Clone not implemented")
}
func (UnimplementedDaemonServer) Remove(*RemoveRequest, Daemon_RemoveServer) error {
	return status.Errorf(codes.Unimplemented, "method Remove not implemented")
}
func (UnimplementedDaemonServer) Logs(*LogRequest, Daemon_LogsServer) error {
	return status.Errorf(codes.Unimplemented, "method Logs not implemented")
}
func (UnimplementedDaemonServer) List(context.Context, *ListRequest) (*ListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (UnimplementedDaemonServer) Upgrade(*UpgradeRequest, Daemon_UpgradeServer) error {
	return status.Errorf(codes.Unimplemented, "method Upgrade not implemented")
}
func (UnimplementedDaemonServer) Status(context.Context, *StatusRequest) (*StatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Status not implemented")
}
func (UnimplementedDaemonServer) Quit(*QuitRequest, Daemon_QuitServer) error {
	return status.Errorf(codes.Unimplemented, "method Quit not implemented")
}
func (UnimplementedDaemonServer) mustEmbedUnimplementedDaemonServer() {}

// UnsafeDaemonServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DaemonServer will
// result in compilation errors.
type UnsafeDaemonServer interface {
	mustEmbedUnimplementedDaemonServer()
}

func RegisterDaemonServer(s grpc.ServiceRegistrar, srv DaemonServer) {
	s.RegisterService(&Daemon_ServiceDesc, srv)
}

func _Daemon_Connect_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ConnectRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(DaemonServer).Connect(m, &daemonConnectServer{stream})
}

type Daemon_ConnectServer interface {
	Send(*ConnectResponse) error
	grpc.ServerStream
}

type daemonConnectServer struct {
	grpc.ServerStream
}

func (x *daemonConnectServer) Send(m *ConnectResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Daemon_Disconnect_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(DisconnectRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(DaemonServer).Disconnect(m, &daemonDisconnectServer{stream})
}

type Daemon_DisconnectServer interface {
	Send(*DisconnectResponse) error
	grpc.ServerStream
}

type daemonDisconnectServer struct {
	grpc.ServerStream
}

func (x *daemonDisconnectServer) Send(m *DisconnectResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Daemon_Proxy_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ConnectRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(DaemonServer).Proxy(m, &daemonProxyServer{stream})
}

type Daemon_ProxyServer interface {
	Send(*ConnectResponse) error
	grpc.ServerStream
}

type daemonProxyServer struct {
	grpc.ServerStream
}

func (x *daemonProxyServer) Send(m *ConnectResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Daemon_Leave_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(LeaveRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(DaemonServer).Leave(m, &daemonLeaveServer{stream})
}

type Daemon_LeaveServer interface {
	Send(*LeaveResponse) error
	grpc.ServerStream
}

type daemonLeaveServer struct {
	grpc.ServerStream
}

func (x *daemonLeaveServer) Send(m *LeaveResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Daemon_Clone_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(CloneRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(DaemonServer).Clone(m, &daemonCloneServer{stream})
}

type Daemon_CloneServer interface {
	Send(*CloneResponse) error
	grpc.ServerStream
}

type daemonCloneServer struct {
	grpc.ServerStream
}

func (x *daemonCloneServer) Send(m *CloneResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Daemon_Remove_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(RemoveRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(DaemonServer).Remove(m, &daemonRemoveServer{stream})
}

type Daemon_RemoveServer interface {
	Send(*RemoveResponse) error
	grpc.ServerStream
}

type daemonRemoveServer struct {
	grpc.ServerStream
}

func (x *daemonRemoveServer) Send(m *RemoveResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Daemon_Logs_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(LogRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(DaemonServer).Logs(m, &daemonLogsServer{stream})
}

type Daemon_LogsServer interface {
	Send(*LogResponse) error
	grpc.ServerStream
}

type daemonLogsServer struct {
	grpc.ServerStream
}

func (x *daemonLogsServer) Send(m *LogResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Daemon_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DaemonServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Daemon_List_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DaemonServer).List(ctx, req.(*ListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Daemon_Upgrade_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(UpgradeRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(DaemonServer).Upgrade(m, &daemonUpgradeServer{stream})
}

type Daemon_UpgradeServer interface {
	Send(*UpgradeResponse) error
	grpc.ServerStream
}

type daemonUpgradeServer struct {
	grpc.ServerStream
}

func (x *daemonUpgradeServer) Send(m *UpgradeResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Daemon_Status_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DaemonServer).Status(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Daemon_Status_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DaemonServer).Status(ctx, req.(*StatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Daemon_Quit_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(QuitRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(DaemonServer).Quit(m, &daemonQuitServer{stream})
}

type Daemon_QuitServer interface {
	Send(*QuitResponse) error
	grpc.ServerStream
}

type daemonQuitServer struct {
	grpc.ServerStream
}

func (x *daemonQuitServer) Send(m *QuitResponse) error {
	return x.ServerStream.SendMsg(m)
}

// Daemon_ServiceDesc is the grpc.ServiceDesc for Daemon service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Daemon_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "rpc.Daemon",
	HandlerType: (*DaemonServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "List",
			Handler:    _Daemon_List_Handler,
		},
		{
			MethodName: "Status",
			Handler:    _Daemon_Status_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Connect",
			Handler:       _Daemon_Connect_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Disconnect",
			Handler:       _Daemon_Disconnect_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Proxy",
			Handler:       _Daemon_Proxy_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Leave",
			Handler:       _Daemon_Leave_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Clone",
			Handler:       _Daemon_Clone_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Remove",
			Handler:       _Daemon_Remove_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Logs",
			Handler:       _Daemon_Logs_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Upgrade",
			Handler:       _Daemon_Upgrade_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Quit",
			Handler:       _Daemon_Quit_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "daemon.proto",
}
