package test_setup

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/anonopiran/Fly2User/internal/config"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

type DeploymentStruct struct {
	HostName         string
	ImageName        string
	ConfigFileSrc    string
	ConfigFileTarget string
	List             []string
	Cmd              strslice.StrSlice
	Ips              mapset.Set[string]
	startMsg         string
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
	hst_cfg := container.HostConfig{AutoRemove: true, NetworkMode: network.NetworkBridge}
	if d.ConfigFileSrc != "" && d.ConfigFileTarget != "" {
		hst_cfg.Binds = []string{fmt.Sprintf("%s/%s:%s", dir, d.ConfigFileSrc, d.ConfigFileTarget)}
	}
	startWg := sync.WaitGroup{}
	for range count {
		resp, err := cli.ContainerCreate(ctx, &cntiner_cfg, &hst_cfg, nil, nil, "")
		if err != nil {
			panic(err)
		}
		cntinerID := resp.ID
		if err := cli.ContainerStart(ctx, cntinerID, container.StartOptions{}); err != nil {
			panic(err)
		}
		startWg.Add(1)
		go waitForLogMessage(ctx, cli, cntinerID, d.startMsg, &startWg)
		d.List = append(d.List, cntinerID)
		containerMeta, err := cli.ContainerInspect(ctx, cntinerID)
		if err != nil {
			panic(err)
		}
		d.Ips.Add(containerMeta.NetworkSettings.IPAddress)
	}
	startWg.Wait()
	time.Sleep(1 * time.Second)
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
	time.Sleep(1 * time.Second)
}
func (d *DeploymentStruct) Restart() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()
	noWaitTimeout := 0
	d.Ips = mapset.NewSet[string]()
	wg := sync.WaitGroup{}
	for _, cntinerID := range d.List {
		if err := cli.ContainerRestart(ctx, cntinerID, container.StopOptions{Timeout: &noWaitTimeout}); err != nil {
			panic(err)
		}
		wg.Add(1)
		go waitForLogMessage(ctx, cli, cntinerID, d.startMsg, &wg)
		containerMeta, err := cli.ContainerInspect(ctx, cntinerID)
		if err != nil {
			panic(err)
		}
		d.Ips.Add(containerMeta.NetworkSettings.IPAddress)
	}
	wg.Wait()
	time.Sleep(1 * time.Second)
}

func NewV2FlyServerDeployment() *DeploymentStruct {
	return &DeploymentStruct{
		HostName:         "v2fly-server.fly2user.local",
		ImageName:        "v2fly/v2fly-core:v5.16.1",
		ConfigFileSrc:    "v2ray-server.json",
		ConfigFileTarget: "/etc/v2ray/config.json",
		Cmd:              strslice.StrSlice{"run", "-c", "/etc/v2ray/config.json"},
		List:             []string{},
		Ips:              mapset.NewSet[string](),
		startMsg:         "V2Ray 5.16.1 started",
	}
}
func NewXrayServerDeployment() *DeploymentStruct {
	return &DeploymentStruct{
		HostName:         "xray-server.fly2user.local",
		ImageName:        "teddysun/xray:1.8.23",
		ConfigFileSrc:    "v2ray-server.json",
		ConfigFileTarget: "/etc/xray/config.json",
		Cmd:              strslice.StrSlice{},
		List:             []string{},
		Ips:              mapset.NewSet[string](),
		startMsg:         "Xray 1.8.23 started",
	}
}

// ...
func waitForLogMessage(ctx context.Context, cli *client.Client, containerID string, targetMessage string, wg *sync.WaitGroup) {
	defer wg.Done()
	if targetMessage == "" {
		return
	}
	logChan := make(chan struct{})
	errChan := make(chan error)
	defer close(logChan)
	defer close(errChan)
	go func() {
		logs, err := cli.ContainerLogs(ctx, containerID, container.LogsOptions{
			ShowStdout: true,
			Follow:     true,
		})
		if err != nil {
			errChan <- err
		}
		defer logs.Close()

		buf := make([]byte, 1024)
		for {
			n, err := logs.Read(buf)
			if err != nil {
				if err != io.EOF {
					errChan <- err
				}
				break
			}

			logLine := string(buf[:n])
			if strings.Contains(logLine, targetMessage) {
				logChan <- struct{}{}
				break
			}
		}
	}()

	select {
	case <-logChan:
		return
	case err := <-errChan:
		logrus.WithError(err).Error("error waiting for container")
		return
	case <-time.After(5 * time.Second): // Adjust timeout as needed
		logrus.Error("timeout waiting for container")
		return
	}
}

// ...
var INBOUNDS = []config.InboundConfigType{
	// {Proto: config.VMESS_PROTO, Tag: "ix_vmess"},
	{Proto: config.VMESS_PROTO, Tag: "i_vmess"},
	{Proto: config.VLESS_PROTO, Tag: "i_vless"},
	{Proto: config.TROJAN_PROTO, Tag: "i_trj"},
}
