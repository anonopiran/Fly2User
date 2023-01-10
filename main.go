package main

import (
	api "Fly2User/Api2User"
	supervisor "Fly2User/Supervisor"
)

func main() {
	go supervisor.Supervise()
	api.Serve()
}
