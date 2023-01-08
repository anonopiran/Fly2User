package api2user

import (
	"net/http"

	config "Fly2User/Config"
	supervisor "Fly2User/Supervisor"
	u2g "Fly2User/User2Grpc"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/hashstructure/v2"
	log "github.com/sirupsen/logrus"
)

var cfg = config.Config()

// ==================================
// functions
// ==================================
func add_user() gin.HandlerFunc {
	return func(c *gin.Context) {
		uar := u2g.UserAddType{}
		if err := c.BindJSON(&uar); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		email := uar.Email
		if supervisor.UserCheckFile(email) {
			existing, err := supervisor.UserFromFile(email)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			req_hash, err := hashstructure.Hash(uar, hashstructure.FormatV2, nil)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			existing_hash, err := hashstructure.Hash(existing, hashstructure.FormatV2, nil)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			if req_hash != existing_hash {
				c.AbortWithStatusJSON(http.StatusConflict, map[string]string{"msg": "user with same email and different config already exists"})
				return
			}
		}
		err := supervisor.UserToFile(uar)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		// ........
		conn, err := u2g.NewGrpcConn(cfg.V2flyApiAddress)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			supervisor.UserDelFile(email)
			return
		}
		defer conn.Close()

		// err_list := u2g.AddUser(conn, cfg.InboundList, &uar, false)
		err_list := uar.AddMultiple(conn, &cfg.InboundList, false)
		errs_res := []string{}
		for _, err = range err_list {
			errs_res = append(errs_res, err.Error())
		}
		if len(err_list) > 0 {
			urr := u2g.UserRemoveType{Email: email}
			urr.RemoveMultiple(conn, &cfg.InboundList, false)
			supervisor.UserDelFile(email)
			c.AbortWithStatusJSON(http.StatusBadRequest, map[string][]string{"msg": errs_res})
			return
		}
		c.AbortWithStatusJSON(http.StatusAccepted, uar)
	}
}
func remove_user() gin.HandlerFunc {
	return func(c *gin.Context) {
		urr := u2g.UserRemoveType{}
		if err := c.BindJSON(&urr); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		email := urr.Email
		// ........
		conn, err := u2g.NewGrpcConn(cfg.V2flyApiAddress)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer conn.Close()
		err_list := urr.RemoveMultiple(conn, &cfg.InboundList, false)
		errs_res := []string{}
		for _, err = range err_list {
			errs_res = append(errs_res, err.Error())
		}
		if len(err_list) > 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, map[string][]string{"msg": errs_res})
			return
		}
		if err := supervisor.UserDelFile(email); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.AbortWithStatus(http.StatusAccepted)
	}
}

func Serve() {
	app := gin.Default()
	api_user := app.Group("/user")
	{
		api_user.POST("/", add_user())
		api_user.DELETE("/", remove_user())
	}
	log.Fatal(app.Run(":3000"))
}
