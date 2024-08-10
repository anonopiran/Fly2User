package v2ray

import (
	"fmt"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ...
type GrpcError struct {
	Err error
}

func (err *GrpcError) Error() string {
	return err.Err.Error()
}
func (err *GrpcError) IsUserExistsError() bool {

	return err.Err != nil && strings.HasSuffix(err.Error(), "already exists.")
}
func (err *GrpcError) IsUserNotFoundError() bool {
	return err.Err != nil && strings.HasSuffix(err.Error(), "not found.")
}

// ...
func NewInsecureGrpc(address net.IP, port string) (*grpc.ClientConn, error) {
	if address == nil {
		return nil, ErrNilAddress
	}
	if port == "" {
		return nil, ErrNilPort
	}
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", address.String(), port), opt)
	if err != nil {
		return nil, fmt.Errorf("error dial %s:%s:%s", address.String(), port, err)
	}
	return conn, nil
}
