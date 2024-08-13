package test_setup

import (
	mapset "github.com/deckarep/golang-set/v2"
)

type ConnTestStruct struct {
	DeploymentStruct
}

func (d *ConnTestStruct) Deploy() {
	d.DeploymentStruct.Deploy(1)
}
func (d *ConnTestStruct) UnDeploy() {
	d.DeploymentStruct.UnDeploy()
}
func NewConnTestDeployment() *ConnTestStruct {
	return &ConnTestStruct{DeploymentStruct{
		HostName:  "conn.fly2user.local",
		ImageName: "containous/whoami:latest",
		List:      []string{},
		Ips:       mapset.NewSet[string](),
		startMsg:  "Starting up on port 80",
	},
	}
}
