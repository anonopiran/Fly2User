package test_setup

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/anonopiran/Fly2User/internal/config"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
)

type DeploymentStruct struct {
	HostName         string
	ImageName        string
	ConfigFileSrc    string
	ConfigFileTarget string
	List             []string
	Cmd              strslice.StrSlice
	Ips              mapset.Set[string]
}

func (d *DeploymentStruct) Deploy(count int) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()
	// ...
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Could not get the caller information")
	}
	dir := filepath.Dir(filename)
	// ...
	if _, err = cli.ImagePull(ctx, d.ImageName, image.PullOptions{}); err != nil {
		panic(err)
	}
	cntiner_cfg := container.Config{Image: d.ImageName, Hostname: d.HostName, Cmd: d.Cmd}
	hst_cfg := container.HostConfig{
		AutoRemove: true,
		Binds:      []string{fmt.Sprintf("%s/%s:%s", dir, d.ConfigFileSrc, d.ConfigFileTarget)},
	}
	for range count {
		resp, err := cli.ContainerCreate(ctx, &cntiner_cfg, &hst_cfg, nil, nil, "")
		if err != nil {
			panic(err)
		}
		cntinerID := resp.ID
		if err := cli.ContainerStart(ctx, cntinerID, container.StartOptions{}); err != nil {
			panic(err)
		}
		d.List = append(d.List, cntinerID)
		containerMeta, err := cli.ContainerInspect(ctx, cntinerID)
		if err != nil {
			panic(err)
		}
		d.Ips.Add(containerMeta.NetworkSettings.IPAddress)
	}
	time.Sleep(2 * time.Second)
}
func (d *DeploymentStruct) UnDeploy() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()
	noWaitTimeout := 0
	for _, cntID := range d.List {
		if err := cli.ContainerStop(ctx, cntID, container.StopOptions{Timeout: &noWaitTimeout}); err != nil {
			panic(err)
		}
	}
	d.List = []string{}
	d.Ips = mapset.NewSet[string]()
}
func (d *DeploymentStruct) Restart() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()
	noWaitTimeout := 0
	for _, cntID := range d.List {
		if err := cli.ContainerRestart(ctx, cntID, container.StopOptions{Timeout: &noWaitTimeout}); err != nil {
			panic(err)
		}
	}
	d.List = []string{}
}

func NewV2FlyServerDeployment() *DeploymentStruct {
	return &DeploymentStruct{
		HostName:         "v2fly-server.fly2user.local",
		ImageName:        "v2fly/v2fly-core:latest",
		ConfigFileSrc:    "v2ray-server.json",
		ConfigFileTarget: "/etc/v2ray/config.json",
		Cmd:              strslice.StrSlice{"run", "-c", "/etc/v2ray/config.json"},
		List:             []string{},
		Ips:              mapset.NewSet[string](),
	}
}
func NewV2FlyClientDeployment() *DeploymentStruct {
	return &DeploymentStruct{
		HostName:         "v2fly-client.fly2user.local",
		ImageName:        "v2fly/v2fly-core:latest",
		ConfigFileSrc:    "v2ray-client.json",
		ConfigFileTarget: "/etc/v2ray/config.json",
		Cmd:              strslice.StrSlice{"run", "-c", "/etc/v2ray/config.json"},
		List:             []string{},
		Ips:              mapset.NewSet[string](),
	}
}
func NewXrayServerDeployment() *DeploymentStruct {
	return &DeploymentStruct{
		HostName:         "xray-server.fly2user.local",
		ImageName:        "teddysun/xray:latest",
		ConfigFileSrc:    "v2ray-server.json",
		ConfigFileTarget: "/etc/xray/config.json",
		Cmd:              strslice.StrSlice{},
		List:             []string{},
		Ips:              mapset.NewSet[string](),
	}
}
func NewXrayClientDeployment() *DeploymentStruct {
	return &DeploymentStruct{
		HostName:         "xray-client.fly2user.local",
		ImageName:        "teddysun/xray:latest",
		ConfigFileSrc:    "v2ray-client.json",
		ConfigFileTarget: "/etc/xray/config.json",
		Cmd:              strslice.StrSlice{},
		List:             []string{},
		Ips:              mapset.NewSet[string](),
	}
}

var INBOUNDS = []config.InboundConfigType{
	{Proto: config.VMESS_PROTO, Tag: "i_vmess"},
	{Proto: config.VLESS_PROTO, Tag: "i_vless"},
	{Proto: config.TROJAN_PROTO, Tag: "i_trj"},
}
