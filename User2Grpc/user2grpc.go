package user2grpc

import (
	"context"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/any"
	proxy "github.com/v2fly/v2ray-core/v5/app/proxyman/command"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/proxy/trojan"
	"github.com/v2fly/v2ray-core/v5/proxy/vless"
	"github.com/v2fly/v2ray-core/v5/proxy/vmess"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/runtime/protoiface"
	"google.golang.org/protobuf/types/known/anypb"
)

// ==================================
// types
// ==================================
type PROTO int

const (
	VMESS_PROTO PROTO = iota
	VLESS_PROTO
	TROJAN_PROTO
	NULL_PROTO
)

type InboundType struct {
	Tag   string
	Proto PROTO
}

type UserAddType struct {
	Uuid  string
	Email string
	Level uint32
}
type UserRemoveType struct {
	Email string
}

// ==================================
// functions
// ==================================
func GrpcClient(server string) (*grpc.ClientConn, error) {
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(server, opt)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func _add_user(conn *grpc.ClientConn, inb *InboundType, user *UserAddType, exists_err bool) error {
	var account *any.Any
	var err error
	var acc protoiface.MessageV1

	switch inb.Proto {
	case VMESS_PROTO:
		acc = &vmess.Account{Id: user.Uuid}
	case VLESS_PROTO:
		acc = &vless.Account{Id: user.Uuid}
	case TROJAN_PROTO:
		acc = &trojan.Account{Password: user.Uuid}
	}
	account = serial.ToTypedMessage(acc)

	user_ob := protocol.User{Account: account, Email: user.Email, Level: user.Level}
	add_user_op := proxy.AddUserOperation{User: &user_ob}
	req := inbound_alter_request(inb.Tag, serial.ToTypedMessage(&add_user_op))
	// .....
	client := proxy.NewHandlerServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = client.AlterInbound(ctx, &req)

	if err != nil {
		if !AddUserErrIsExists(err) || exists_err {
			return err
		}
	}
	return nil
}

func _remove_user(conn *grpc.ClientConn, inb InboundType, user *UserRemoveType) error {
	rm_user_op := proxy.RemoveUserOperation{Email: user.Email}
	req := inbound_alter_request(inb.Tag, serial.ToTypedMessage(&rm_user_op))
	// .....
	client := proxy.NewHandlerServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.AlterInbound(ctx, &req)
	if err != nil {
		if strings.HasSuffix(err.Error(), "not found.") {
		} else {
			return err
		}
	}
	return nil
}
func AddUser(conn *grpc.ClientConn, inbs []InboundType, user *UserAddType, exists_err bool) []error {
	err := []error{}
	for _, inb := range inbs {
		e := _add_user(conn, &inb, user, exists_err)
		if e != nil {
			err = append(err, e)
		}
	}
	return err
}
func RemoveUser(conn *grpc.ClientConn, inbs []InboundType, user *UserRemoveType) []error {
	err := []error{}
	for _, inb := range inbs {
		e := _remove_user(conn, inb, user)
		if e != nil {
			err = append(err, e)
		}
	}
	return err
}
func inbound_alter_request(tag string, op *anypb.Any) proxy.AlterInboundRequest {
	return proxy.AlterInboundRequest{Tag: tag, Operation: op}
}
func AddUserErrIsExists(err error) bool {
	return strings.HasSuffix(err.Error(), "already exists.")
}
