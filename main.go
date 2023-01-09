package main

import (
	api "Fly2User/Api2User"
	config "Fly2User/Config"
	supervisor "Fly2User/Supervisor"
)

func main() {
	cfg := config.Config()
	supervisor.SetUp(cfg.V2flyApiAddress, string(cfg.UserDir), cfg.InboundList)
	api.SetUp(cfg.V2flyApiAddress, cfg.InboundList)
	go supervisor.Supervise(cfg.SuperviseInterval, cfg.SuperviseUuid)
	api.Serve()
}
