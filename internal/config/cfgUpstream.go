package config

import (
	"errors"
	"net/url"
	"strings"
)

type UpstreamUrlType struct {
	url.URL
	ServerType ServerTypeEnumType
}
type InboundConfigType struct {
	Proto ProtocolEnumType
	Tag   string
}

type UpstreamConfigType struct {
	Address     []UpstreamUrlType   `koanf:"address" validate:"required,min=1"`
	InboundList []InboundConfigType `koanf:"inbounds" validate:"required,min=1"`
}

// ...
type ServerTypeEnumType string
type ProtocolEnumType string

const (
	V2FLY_SRV ServerTypeEnumType = "v2fly"
	XRAY_SRV  ServerTypeEnumType = "xray"
)
const (
	VMESS_PROTO  ProtocolEnumType = "VMESS"
	VLESS_PROTO  ProtocolEnumType = "VLESS"
	TROJAN_PROTO ProtocolEnumType = "TROJAN"
)

// ...
func (f *UpstreamUrlType) UnmarshalText(text []byte) error {
	u, err := url.Parse(string(text))
	if err != nil {
		return err
	}
	if u.Scheme != "grpc" {
		return errors.New("scheme should be grpc")
	}
	if u.Hostname() == "" {
		return errors.New("hostname not provided")
	}
	if u.Port() == "" {
		return errors.New("port not provided")
	}
	var srvType ServerTypeEnumType
	srvTypeConfig := u.User.Username()
	switch ss := ServerTypeEnumType(srvTypeConfig); ss {
	case V2FLY_SRV, XRAY_SRV:
		srvType = ss
	default:
		return errors.New("servertype not understood")
	}
	u.User = nil
	*f = UpstreamUrlType{URL: *u, ServerType: srvType}
	return nil
}

func (f *InboundConfigType) UnmarshalText(text []byte) error {
	s := string(text)
	if match := strings.Contains(s, ":"); !match {
		return errors.New("error parsing inbound")
	}
	x_ := strings.Split(s, ":")
	tag := strings.Join(x_[:len(x_)-1], ":")
	proto_s := x_[len(x_)-1]
	var proto ProtocolEnumType
	switch ss := ProtocolEnumType(proto_s); ss {
	case VMESS_PROTO, VLESS_PROTO, TROJAN_PROTO:
		proto = ss
	default:
		return errors.New("protocol not understood")
	}
	*f = InboundConfigType{Tag: tag, Proto: proto}
	return nil
}
