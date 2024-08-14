package api

import (
	"net/http"

	"github.com/anonopiran/Fly2User/internal/supervisor"
	"github.com/gin-gonic/gin"
)

func addUserHandler(sup *supervisor.Supervisor) gin.HandlerFunc {
	return func(c *gin.Context) {
		usr := &AddUserReq{}
		if err := c.BindJSON(usr); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if err := sup.AddUser(usr.AsUSer()); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
		}
		c.AbortWithStatusJSON(http.StatusAccepted, usr)
	}
}
func rmUserHandler(sup *supervisor.Supervisor) gin.HandlerFunc {
	return func(c *gin.Context) {
		usr := &RmUserReq{}
		if err := c.BindJSON(&usr); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if err := sup.RmUser(usr.AsUSer()); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		c.AbortWithStatusJSON(http.StatusAccepted, usr)
	}
}

func countUserHandler(sup *supervisor.Supervisor) gin.HandlerFunc {
	return func(c *gin.Context) {
		cnt, err := sup.CountUser()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, cnt)
	}
}
func flushUserHandler(sup *supervisor.Supervisor) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := sup.FlushUser(); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.AbortWithStatus(http.StatusOK)
	}
}
