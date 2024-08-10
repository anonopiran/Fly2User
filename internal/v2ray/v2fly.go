package v2ray

import (
	"context"

	"github.com/anonopiran/Fly2User/internal/config"
	proxyV2fly "github.com/v2fly/v2ray-core/v5/app/proxyman/command"

	protocolV2fly "github.com/v2fly/v2ray-core/v5/common/protocol"
	serialV2fly "github.com/v2fly/v2ray-core/v5/common/serial"
	trojanV2fly "github.com/v2fly/v2ray-core/v5/proxy/trojan"
	vlessV2fly "github.com/v2fly/v2ray-core/v5/proxy/vless"
	vmessV2fly "github.com/v2fly/v2ray-core/v5/proxy/vmess"
	"google.golang.org/grpc"
)

type V2flyServer struct {
	UpServer
}

type v2flyHandlerServiceClientAdapter struct {
	client proxyV2fly.HandlerServiceClient
}

func (a *v2flyHandlerServiceClientAdapter) AlterInbound(ctx context.Context, req IRequest, opts ...grpc.CallOption) (IResponse, error) {
	alterInboundRequest, ok := req.(*proxyV2fly.AlterInboundRequest)
	if !ok {
		return nil, ErrUnknownRequestType(req)
	}
	return a.client.AlterInbound(ctx, alterInboundRequest, opts...)
}

// ...
func (v *V2flyServer) getAccount(inbound *config.InboundConfigType, user *UserType) (IRequest, error) {
	if inbound == nil || inbound.Proto == "" {
		return nil, ErrInboundNotDefined
	}
	if user == nil || user.Secret == "" {
		return nil, ErrUserNotDefined
	}
	var acc IRequest
	switch inbound.Proto {
	case config.VMESS_PROTO:
		acc = &vmessV2fly.Account{Id: user.Secret}
	case config.VLESS_PROTO:
		acc = &vlessV2fly.Account{Id: user.Secret}
	case config.TROJAN_PROTO:
		acc = &trojanV2fly.Account{Password: user.Secret}
	default:
		return nil, ErrUnknownInboundProto
	}
	return acc, nil
}
func (v *V2flyServer) NewAddUserReq(inbound *config.InboundConfigType, user *UserType) (IRequest, error) {
	if inbound == nil {
		return nil, ErrInboundNotDefined
	}
	if user == nil {
		return nil, ErrUserNotDefined
	}
	acc, err := v.getAccount(inbound, user)
	if err != nil {
		return nil, err
	}
	usr := protocolV2fly.User{Level: user.Level, Email: user.Email, Account: serialV2fly.ToTypedMessage(acc)}
	addUserOp := proxyV2fly.AddUserOperation{User: &usr}
	alterInboundReq := &proxyV2fly.AlterInboundRequest{Tag: inbound.Tag, Operation: serialV2fly.ToTypedMessage(&addUserOp)}
	return alterInboundReq, nil
}
func (v *V2flyServer) NewRmUserReq(inbound *config.InboundConfigType, user *UserType) (IRequest, error) {
	if inbound == nil {
		return nil, ErrInboundNotDefined
	}
	if user == nil {
		return nil, ErrUserNotDefined
	}
	rmUserOp := proxyV2fly.RemoveUserOperation{Email: user.Email}
	alterInboundReq := &proxyV2fly.AlterInboundRequest{Tag: inbound.Tag, Operation: serialV2fly.ToTypedMessage(&rmUserOp)}
	return alterInboundReq, nil
}
func (v *V2flyServer) execAlterInbound(handler HandlerServiceClient, in IRequest, ctx context.Context, opts ...grpc.CallOption) (IResponse, error) {
	if handler == nil {
		return nil, ErrUnknownHandlerClientType(handler)
	}
	if in == nil {
		return nil, ErrUnknownRequestType(in)
	}
	resp, err := handler.AlterInbound(ctx, in, opts...)
	if err != nil {
		return nil, &GrpcError{err}
	}
	return resp, nil
}
func (v *V2flyServer) newHandlerClient(conn *grpc.ClientConn) HandlerServiceClient {
	hndlr := proxyV2fly.NewHandlerServiceClient(conn)
	return &v2flyHandlerServiceClientAdapter{hndlr}
}
