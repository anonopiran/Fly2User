package user2upstream

import (
	"context"
	"net"
	"net/url"
	"strconv"

	log "github.com/sirupsen/logrus"

	proxy "github.com/v2fly/v2ray-core/v5/app/proxyman/command"
	"google.golang.org/protobuf/types/known/anypb"
)

func ResolveV2FlyServer(url url.URL) ([]ServerType, error) {
	logWithUrl := log.WithField("url", url)
	servers := []ServerType{}
	host := url.Hostname()
	port, err := strconv.ParseInt(url.Port(), 10, 0)
	if err != nil {
		logWithUrl.WithError(err).Error("error parsing port number")
		return nil, err
	}
	ips, err := net.DefaultResolver.LookupIP(context.Background(), "ip4", host)
	if err != nil {
		logWithUrl.WithError(err).Error("can not resolve host")
		return nil, err
	}
	for _, ip_ := range ips {
		servers = append(servers, ServerType{Ip: ip_, Uri: url.Hostname(), Port: port})
	}
	logWithUrl.Debugf("discovered %d services", len(servers))
	logWithUrl.Debugf("%v", servers)
	return servers, nil
}

func NewInboundAlterRequest(tag string, op *anypb.Any) proxy.AlterInboundRequest {
	return proxy.AlterInboundRequest{Tag: tag, Operation: op}
}
