package v2ray

import (
	"context"
	"net"

	"github.com/anonopiran/Fly2User/internal/config"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type IServer interface {
	NewAddUserReq(inbound *config.InboundConfigType, user *UserType) (IRequest, error)
	NewRmUserReq(inbound *config.InboundConfigType, user *UserType) (IRequest, error)
	execAlterInbound(handler HandlerServiceClient, in IRequest, ctx context.Context, opts ...grpc.CallOption) (IResponse, error)
	newHandlerClient(conn *grpc.ClientConn) HandlerServiceClient
}
type UpServer struct {
	IServer
	Address config.UpstreamUrlType
}

// ...
func (v *UpServer) Discover(ctx context.Context) (mapset.Set[string], error) {
	ll := v.Logger(nil)
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip4", v.Address.Hostname())
	if err != nil {
		return nil, err
	}
	ll.Debugf("found ips %+v", ips)
	ipSet := mapset.NewSet[string]()
	for _, ip := range ips {
		ipSet.Add(ip.String())
	}
	return ipSet, nil
}
func (v *UpServer) AddUser(ctx context.Context, inbound *config.InboundConfigType, user *UserType, conn *grpc.ClientConn) error {
	if conn == nil {
		return ErrNilConnection
	}
	req, err := v.NewAddUserReq(inbound, user)
	if err != nil {
		return err
	}
	return v.alterInbound(conn, req, ctx)
}
func (v *UpServer) RmUser(ctx context.Context, inbound *config.InboundConfigType, user *UserType, conn *grpc.ClientConn) error {
	if conn == nil {
		return ErrNilConnection
	}
	req, err := v.NewRmUserReq(inbound, user)
	if err != nil {
		return err
	}
	return v.alterInbound(conn, req, ctx)
}
func (v *UpServer) alterInbound(conn *grpc.ClientConn, req IRequest, ctx context.Context, opts ...grpc.CallOption) error {
	if req == nil {
		return ErrNilRequest
	}
	handler := v.newHandlerClient(conn)
	_, err := v.execAlterInbound(handler, req, ctx, opts...)
	return err
}
func (v *UpServer) Logger(ll *logrus.Entry) *logrus.Entry {
	if ll == nil {
		ll = logrus.NewEntry(logrus.StandardLogger())
	}
	return ll.WithField("upstream", v.Address.String())
}

// ...
func NewServer(srv config.UpstreamUrlType) (*UpServer, error) {
	var upSrv UpServer
	switch srv.ServerType {
	case config.V2FLY_SRV:
		upSrv = UpServer{Address: srv, IServer: &V2flyServer{}}
	case config.XRAY_SRV:
		upSrv = UpServer{Address: srv, IServer: &XrayServer{}}
	default:
		return nil, ErrUnknownServerType(srv)
	}
	return &upSrv, nil
}
