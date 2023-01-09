package discovery

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ==================================
// types
// ==================================
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

// ==================================
// functions
// ==================================
func ResolveServers(urls []string) []ServerType {
	servers := []ServerType{}
	for _, srv := range urls {
		x := strings.Split(srv, ":")
		host := x[0]
		port, err := strconv.ParseInt(x[1], 10, 0)
		if err != nil {
			log.WithError(err).Panicf("can not parse port %s", srv)
		}
		ips, err := net.DefaultResolver.LookupIP(context.Background(), "ip4", host)
		if err != nil {
			log.WithError(err).Errorf("can not resolve host %s. skiping...", host)
			continue
		}
		for _, ip_ := range ips {
			servers = append(servers, ServerType{Ip: ip_, Uri: host, Port: port})
		}
	}
	log.Debugf("discovered %d services", len(servers))
	log.Debugf("%v", servers)
	return servers
}
