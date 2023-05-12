package api2user

import (
	config "Fly2User/Config"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func Serve() {
	cfg := config.Config()
	app := gin.Default()
	api_user := app.Group("/user")
	{
		api_user.POST("/", addUserHandler())
		api_user.DELETE("/", removeUserHandler())
		api_user.GET("/count", countUserHandler())
	}
	log.Fatal(app.Run(cfg.Listen))
}
