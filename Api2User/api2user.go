package api2user

import (
	config "Fly2User/Config"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func Serve() {
	app := gin.Default()
	api_user := app.Group("/user")
	{
		api_user.POST("/", addUserHandler())
		api_user.DELETE("/", removeUserHandler())
		api_user.GET("/count", countUserHandler())
	}
	log.Fatal(app.Run(config.Config.Listen))
}
func init() {
	servers = config.Config.V2flyApiAddress
	inbounds = config.Config.InboundList
}
