package user2upstream

import (
	"context"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/any"
	log "github.com/sirupsen/logrus"
	proxy "github.com/v2fly/v2ray-core/v5/app/proxyman/command"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/proxy/trojan"
	"github.com/v2fly/v2ray-core/v5/proxy/vless"
	"github.com/v2fly/v2ray-core/v5/proxy/vmess"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/runtime/protoiface"
)

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
type ServerType struct {
	Uri  string
	Ip   net.IP
	Port int64
}

func (srv *ServerType) DialGrpc() (*grpc.ClientConn, error) {
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", srv.Ip.String(), srv.Port), opt)
	return conn, err
}
func (user *UserAddType) Add(conn *grpc.ClientConn, inb *InboundType, existsErr bool) GrpcError {
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
		if !err.IsUserExistsError() || existsErr {
			return err
		}
	}
	return nil
}
func (user *UserRemoveType) Remove(conn *grpc.ClientConn, inb *InboundType, notFoundErr bool) GrpcError {
	rm_user_op := proxy.RemoveUserOperation{Email: user.Email}
	req := NewInboundAlterRequest(inb.Tag, serial.ToTypedMessage(&rm_user_op))
	// .....
	client := proxy.NewHandlerServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.AlterInbound(ctx, &req)
	if err != nil {
		err := &GrpcErrorType{Err: err, Inbound: *inb}
		if !err.IsUserNotFoundError() || notFoundErr {
			return err
		}
	}
	return nil
}
func (user *UserAddType) AddMultiple(conn *grpc.ClientConn, inbs *[]InboundType, existsErr bool) []GrpcError {
	err := []GrpcError{}
	for _, inb := range *inbs {
		e := user.Add(conn, &inb, existsErr)
		if e != nil {
			err = append(err, e)
		}
	}
	return err
}
func (user *UserRemoveType) RemoveMultiple(conn *grpc.ClientConn, inbs *[]InboundType, notFoundErr bool) []GrpcError {
	err := []GrpcError{}
	for _, inb := range *inbs {
		e := user.Remove(conn, &inb, notFoundErr)
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

func (f *InboundType) SetValue(s string) error {
	LogWithRaw := log.WithField("value", s)
	match, _ := regexp.MatchString(".+:(VMESS|VLESS|TROJAN)", s)
	if !match {
		e := errors.New("error parsing inbound")
		LogWithRaw.Error(e)
		return e
	}
	x_ := strings.Split(s, ":")
	tag := x_[0]
	proto_s := x_[1]
	var proto ProtocolType
	switch proto_s {
	case "VMESS":
		proto = VMESS_PROTO
	case "VLESS":
		proto = VLESS_PROTO
	case "TROJAN":
		proto = TROJAN_PROTO
	}
	*f = InboundType{Tag: tag, Proto: proto}
	return nil
}
