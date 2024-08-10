package v2ray

import (
	"context"

	"github.com/anonopiran/Fly2User/internal/config"
	proxyXray "github.com/xtls/xray-core/app/proxyman/command"
	protocolXray "github.com/xtls/xray-core/common/protocol"
	serialXray "github.com/xtls/xray-core/common/serial"
	trojanXray "github.com/xtls/xray-core/proxy/trojan"
	vlessXray "github.com/xtls/xray-core/proxy/vless"
	vmessXray "github.com/xtls/xray-core/proxy/vmess"

	"google.golang.org/grpc"
)

type XrayServer struct {
	UpServer
}

type xrayHandlerServiceClientAdapter struct {
	client proxyXray.HandlerServiceClient
}

func (a *xrayHandlerServiceClientAdapter) AlterInbound(ctx context.Context, req IRequest, opts ...grpc.CallOption) (IResponse, error) {
	alterInboundRequest, ok := req.(*proxyXray.AlterInboundRequest)
	if !ok {
		return nil, ErrUnknownRequestType(req)
	}
	return a.client.AlterInbound(ctx, alterInboundRequest, opts...)
}

// ...
func (v *XrayServer) getAccount(inbound *config.InboundConfigType, user *UserType) (IRequest, error) {
	if inbound == nil || inbound.Proto == "" {
		return nil, ErrInboundNotDefined
	}
	if user == nil || user.Secret == "" {
		return nil, ErrUserNotDefined
	}
	var acc IRequest
	switch inbound.Proto {
	case config.VMESS_PROTO:
		acc = &vmessXray.Account{Id: user.Secret}
	case config.VLESS_PROTO:
		acc = &vlessXray.Account{Id: user.Secret}
	case config.TROJAN_PROTO:
		acc = &trojanXray.Account{Password: user.Secret}
	default:
		return nil, ErrUnknownInboundProto
	}
	return acc, nil
}
func (v *XrayServer) NewAddUserReq(inbound *config.InboundConfigType, user *UserType) (IRequest, error) {
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
	usr := protocolXray.User{Level: user.Level, Email: user.Email, Account: serialXray.ToTypedMessage(acc)}
	addUserOp := proxyXray.AddUserOperation{User: &usr}
	alterInboundReq := &proxyXray.AlterInboundRequest{Tag: inbound.Tag, Operation: serialXray.ToTypedMessage(&addUserOp)}
	return alterInboundReq, nil
}
func (v *XrayServer) NewRmUserReq(inbound *config.InboundConfigType, user *UserType) (IRequest, error) {
	if inbound == nil {
		return nil, ErrInboundNotDefined
	}
	if user == nil {
		return nil, ErrUserNotDefined
	}
	rmUserOp := proxyXray.RemoveUserOperation{Email: user.Email}
	alterInboundReq := &proxyXray.AlterInboundRequest{Tag: inbound.Tag, Operation: serialXray.ToTypedMessage(&rmUserOp)}
	return alterInboundReq, nil
}
func (v *XrayServer) execAlterInbound(handler HandlerServiceClient, in IRequest, ctx context.Context, opts ...grpc.CallOption) (IResponse, error) {
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
func (v *XrayServer) newHandlerClient(conn *grpc.ClientConn) HandlerServiceClient {
	hndlr := proxyXray.NewHandlerServiceClient(conn)
	return &xrayHandlerServiceClientAdapter{hndlr}
}
