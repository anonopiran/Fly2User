package user2grpc

import (
	"context"
	"fmt"
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
type ProtocolType int
type GrpcErrorType struct {
	Err     error
	Inbound InboundType
}

const (
	VMESS_PROTO ProtocolType = iota
	VLESS_PROTO
	TROJAN_PROTO
	NULL_PROTO
)

type InboundType struct {
	Tag   string
	Proto ProtocolType
}

type UserAddType struct {
	Uuid  string
	Email string
	Level uint32
}
type UserRemoveType struct {
	Email string
}
type GrpcError interface {
	IsUserExistsError() bool
	IsUserNotFoundError() bool
	GetInbound() InboundType
	Error() string
}

func (f *InboundType) SetValue(s string) error {
	x_ := strings.Split(s, ":")
	tag := x_[0]
	proto_s := x_[1]
	var proto ProtocolType
	switch strings.ToUpper(proto_s) {
	case "VMESS":
		proto = VMESS_PROTO
	case "VLESS":
		proto = VLESS_PROTO
	case "TROJAN":
		proto = TROJAN_PROTO
	default:
		return fmt.Errorf("proto %s not found", s)
	}
	*f = InboundType{Tag: tag, Proto: proto}
	return nil
}
func (user *UserAddType) Add(conn *grpc.ClientConn, inb *InboundType, exists_err bool) GrpcError {
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
	req := NewInboundAlterRequest(inb.Tag, serial.ToTypedMessage(&add_user_op))
	// .....
	client := proxy.NewHandlerServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = client.AlterInbound(ctx, &req)

	if err != nil {
		err := &GrpcErrorType{Err: err, Inbound: *inb}
		if !err.IsUserExistsError() || exists_err {
			return err
		}
	}
	return nil
}

func (user *UserRemoveType) Remove(conn *grpc.ClientConn, inb *InboundType, notfound_err bool) GrpcError {
	rm_user_op := proxy.RemoveUserOperation{Email: user.Email}
	req := NewInboundAlterRequest(inb.Tag, serial.ToTypedMessage(&rm_user_op))
	// .....
	client := proxy.NewHandlerServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.AlterInbound(ctx, &req)
	if err != nil {
		err := &GrpcErrorType{Err: err, Inbound: *inb}
		if !err.IsUserNotFoundError() || notfound_err {
			return err
		}
	}
	return nil
}
func (user *UserAddType) AddMultiple(conn *grpc.ClientConn, inbs *[]InboundType, exists_err bool) []GrpcError {
	err := []GrpcError{}
	for _, inb := range *inbs {
		e := user.Add(conn, &inb, exists_err)
		if e != nil {
			err = append(err, e)
		}
	}
	return err
}
func (user *UserRemoveType) RemoveMultiple(conn *grpc.ClientConn, inbs *[]InboundType, notfound_err bool) []GrpcError {
	err := []GrpcError{}
	for _, inb := range *inbs {
		e := user.Remove(conn, &inb, notfound_err)
		if e != nil {
			err = append(err, e)
		}
	}
	return err
}
func (err *GrpcErrorType) IsUserExistsError() bool {
	return strings.HasSuffix(err.Err.Error(), "already exists.")
}

func (err *GrpcErrorType) IsUserNotFoundError() bool {
	return strings.HasSuffix(err.Err.Error(), "not found.")
}
func (err *GrpcErrorType) GetInbound() InboundType {
	return err.Inbound
}
func (err *GrpcErrorType) Error() string {
	return err.Err.Error()
}

// ==================================
// functions
// ==================================
func NewGrpcConn(server string) (*grpc.ClientConn, GrpcError) {
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(server, opt)
	if err != nil {
		return nil, &GrpcErrorType{Err: err}
	}
	return conn, nil
}
func NewInboundAlterRequest(tag string, op *anypb.Any) proxy.AlterInboundRequest {
	return proxy.AlterInboundRequest{Tag: tag, Operation: op}
}
