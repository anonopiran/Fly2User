package test_setup

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/anonopiran/Fly2User/internal/config"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/docker/docker/api/types/strslice"
	v5 "github.com/v2fly/v2ray-core/v5"
	proxyV2fly "github.com/v2fly/v2ray-core/v5/app/proxyman/command"
	"github.com/v2fly/v2ray-core/v5/common/net"
	protocol "github.com/v2fly/v2ray-core/v5/common/protocol"
	serialV2fly "github.com/v2fly/v2ray-core/v5/common/serial"
	trojan "github.com/v2fly/v2ray-core/v5/proxy/trojan"
	vless "github.com/v2fly/v2ray-core/v5/proxy/vless"
	vlessOutB "github.com/v2fly/v2ray-core/v5/proxy/vless/outbound"
	vmess "github.com/v2fly/v2ray-core/v5/proxy/vmess"
	vmessOutB "github.com/v2fly/v2ray-core/v5/proxy/vmess/outbound"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientStruct struct {
	DeploymentStruct
}

func (d *ClientStruct) Deploy() {
	d.DeploymentStruct.Deploy(1)
}
func (d *ClientStruct) UnDeploy() {
	d.DeploymentStruct.UnDeploy()
}
func (d *ClientStruct) Restart() {
	d.DeploymentStruct.Restart()
}
func (d *ClientStruct) SetOutbound(ip string, id string) error {
	ctx := context.Background()
	ipAdd := net.NewIPOrDomain(net.ParseAddress(ip))
	conn, err := grpc.NewClient(fmt.Sprintf("%s:8080", d.DeploymentStruct.HostName), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()
	hndlr := proxyV2fly.NewHandlerServiceClient(conn)
	for _, outb := range INBOUNDS {
		var req *v5.OutboundHandlerConfig
		switch outb.Proto {
		case config.VMESS_PROTO:
			req = &v5.OutboundHandlerConfig{
				Tag: outb.Tag,
				ProxySettings: serialV2fly.ToTypedMessage(&vmessOutB.Config{
					Receiver: []*protocol.ServerEndpoint{
						{
							Address: ipAdd, Port: 1110, User: []*protocol.User{
								{Email: "test@test.com", Level: 0, Account: serialV2fly.ToTypedMessage(
									&vmess.Account{
										Id: id,
									},
								),
								},
							},
						},
					},
				}),
			}
		case config.VLESS_PROTO:
			req = &v5.OutboundHandlerConfig{
				Tag: outb.Tag,
				ProxySettings: serialV2fly.ToTypedMessage(&vlessOutB.Config{
					Vnext: []*protocol.ServerEndpoint{
						{
							Address: ipAdd, Port: 1111, User: []*protocol.User{
								{Email: "test@test.com", Level: 0, Account: serialV2fly.ToTypedMessage(
									&vless.Account{
										Id: id,
									},
								),
								},
							},
						},
					},
				}),
			}
		case config.TROJAN_PROTO:
			req = &v5.OutboundHandlerConfig{
				Tag: outb.Tag,
				ProxySettings: serialV2fly.ToTypedMessage(&trojan.ClientConfig{
					Server: []*protocol.ServerEndpoint{
						{
							Address: ipAdd, Port: 1112, User: []*protocol.User{
								{Email: "test@test.com", Level: 0, Account: serialV2fly.ToTypedMessage(
									&trojan.Account{
										Password: id,
									},
								),
								},
							},
						},
					},
				}),
			}
		default:
			return fmt.Errorf("not understood")
		}
		rRemove := proxyV2fly.RemoveOutboundRequest{Tag: outb.Tag}
		rAdd := proxyV2fly.AddOutboundRequest{Outbound: req}
		respR, err := hndlr.RemoveOutbound(ctx, &rRemove)
		if err != nil {
			return fmt.Errorf("err rm: (%+v) %s", outb, err)
		}
		if respR.String() != "" {
			return fmt.Errorf("err rm message: %s", err)
		}
		respA, err := hndlr.AddOutbound(ctx, &rAdd)
		if err != nil {
			return fmt.Errorf("err add: %s", err)
		}
		if respA.String() != "" {
			return fmt.Errorf("err add msg: %s", err)
		}
	}
	return nil
}
func (d *ClientStruct) TestURL(expectErr bool, u string) error {
	inbds := []struct {
		port int
		name string
	}{
		{
			port: 10800,
			name: "i_vmess",
		}, {
			port: 10801,
			name: "i_vless",
		}, {
			port: 10802,
			name: "i_trj",
		},
	}
	for _, inb := range inbds {
		err := d.testHttpConnection(u, inb.port)
		if err == nil && expectErr {
			return fmt.Errorf("unexpected success(%s:%d)", d.DeploymentStruct.HostName, inb.port)
		}
		if err != nil && !expectErr {
			return fmt.Errorf("unexpected fail(%s:%d): %s", d.DeploymentStruct.HostName, inb.port, err)
		}
	}
	return nil
}

func (d *ClientStruct) testHttpConnection(u string, proxyPort int) error {
	pUrl, err := url.Parse(fmt.Sprintf("http://%s:%d", d.DeploymentStruct.HostName, proxyPort))
	if err != nil {
		return err
	}
	client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(pUrl), ResponseHeaderTimeout: 500 * time.Millisecond}}
	r, err := client.Get(u)
	if err != nil {
		return err
	}
	if r.StatusCode != 200 {
		return fmt.Errorf("error status code: %d", r.StatusCode)
	}
	return nil

}
func NewV2FlyClientDeployment() *ClientStruct {
	return &ClientStruct{DeploymentStruct{
		HostName:         "v2fly-client.fly2user.local",
		ImageName:        "v2fly/v2fly-core:v5.16.1",
		ConfigFileSrc:    "v2ray-client.json",
		ConfigFileTarget: "/etc/v2ray/config.json",
		Cmd:              strslice.StrSlice{"run", "-c", "/etc/v2ray/config.json"},
		List:             []string{},
		Ips:              mapset.NewSet[string](),
		startMsg:         "V2Ray 5.16.1 started",
	},
	}
}
