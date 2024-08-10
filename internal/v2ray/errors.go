package v2ray

import (
	"errors"
	"fmt"
)

var ErrUserNotDefined = errors.New("user.Secret should be defined")
var ErrInboundNotDefined = errors.New("inbound.Proto should be defined")
var ErrUnknownInboundProto = errors.New("inbound.Proto should be defined")
var ErrNilAddress = errors.New("address cannot be nil")
var ErrNilPort = errors.New("port cannot be empty")
var ErrNilRequest = errors.New("request is nil")
var ErrNilConnection = errors.New("connection is nil")

func ErrUnknownRequestType(req interface{}) error {
	return fmt.Errorf("unknown request type: %T", req)
}
func ErrUnknownServerType(req interface{}) error {
	return fmt.Errorf("unknown server type: %T", req)
}
func ErrUnknownHandlerClientType(req interface{}) error {
	return fmt.Errorf("unknown handler client type: %T", req)
}
