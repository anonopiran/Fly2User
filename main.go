package main

import (
	api "Fly2User/Api2User"
	supervisor "Fly2User/Supervisor"
)

// var cfg = config.Config()

func main() {
	go supervisor.Supervise()
	api.Serve()
}
