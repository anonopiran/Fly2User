package api

import (
	"github.com/anonopiran/Fly2User/internal/config"
	"github.com/anonopiran/Fly2User/internal/supervisor"
	"github.com/gin-gonic/gin"
)

func NewApp(cfg *config.ServerConfigType, sup *supervisor.Supervisor) *gin.Engine {
	app := gin.Default()
	api_user := app.Group("/user")
	if (cfg.Auth != config.AuthType{}) {
		acc := gin.Accounts{}
		acc[cfg.Auth.Username] = cfg.Auth.Password
		api_user.Handlers = append(api_user.Handlers, gin.BasicAuth(acc))
	}
	{
		api_user.POST("/", addUserHandler(sup))
		api_user.DELETE("/", rmUserHandler(sup))
		api_user.GET("/count", countUserHandler(sup))
		api_user.GET("/flush", flushUserHandler(sup))
	}
	return app
}
