package v2ray

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type UserType struct {
	Email  string
	Secret string
	Level  uint32
}

type IRequest interface {
	Descriptor() ([]byte, []int)
	ProtoMessage()
	ProtoReflect() protoreflect.Message
	Reset()
	String() string
}
type IResponse interface {
	Descriptor() ([]byte, []int)
	ProtoMessage()
	ProtoReflect() protoreflect.Message
	Reset()
	String() string
}

type HandlerServiceClient interface {
	AlterInbound(ctx context.Context, in IRequest, opts ...grpc.CallOption) (IResponse, error)
}
