package api2user

import (
	"errors"
	"net/http"

	discovery "Fly2User/Discovery"
	supervisor "Fly2User/Supervisor"
	u2g "Fly2User/User2Grpc"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/hashstructure/v2"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var servers []string
var inbounds []u2g.InboundType

// ==================================
// functions
// ==================================
func SetUp(setServers []string, setInbounds []u2g.InboundType) {
	servers = setServers
	inbounds = setInbounds
}
func manUser(fn func(*grpc.ClientConn, *[]u2g.InboundType, bool) []u2g.GrpcError, exitOnErr bool) bool {
	hasErr := false
	for _, srv := range discovery.ResolveServers(servers) {
		logWithServerData := log.WithField("server", srv)
		conn, err := srv.DialGrpc()
		if err != nil {
			hasErr = true
			logWithServerData.WithError(err).Error("error while dialing grpc")
			if exitOnErr {
				return hasErr
			}
			continue
		}
		defer conn.Close()
		errList := fn(conn, &inbounds, false)
		if len(errList) > 0 {
			for _, err := range errList {
				logWithServerData.WithError(err).Error("error happened while add/remove user")
			}
			hasErr = true
			if exitOnErr {
				return hasErr
			}
		}
	}
	return hasErr
}
func addUser(user *u2g.UserAddType, exitOnErr bool) bool {
	return manUser(user.AddMultiple, exitOnErr)
}
func rmUser(user *u2g.UserRemoveType, exitOnErr bool) bool {
	return manUser(user.RemoveMultiple, exitOnErr)
}
func addUserHandler() gin.HandlerFunc {
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

		// ........
		urr := u2g.UserRemoveType{Email: uar.Email}
		hasErr := addUser(&uar, true)
		if hasErr {
			rmUser(&urr, false)
			c.AbortWithError(http.StatusInternalServerError, errors.New("an unknown error happened while calling upstream server"))
			return
		}
		// ........
		err := supervisor.UserToFile(uar)
		if err != nil {
			rmUser(&urr, false)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.AbortWithStatusJSON(http.StatusAccepted, uar)
	}
}
func removeUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		urr := u2g.UserRemoveType{}
		if err := c.BindJSON(&urr); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		email := urr.Email
		// ........
		uar, urrErr := supervisor.UserFromFile(email)
		hasErr := rmUser(&urr, true)
		if hasErr {
			if urrErr == nil {
				addUser(&uar, false)
			}
			c.AbortWithError(http.StatusInternalServerError, errors.New("an unknown error happened while calling upstream server"))
			return
		}
		// ........
		err := supervisor.UserDelFile(email)
		if err != nil {
			if urrErr == nil {
				addUser(&uar, false)
			}
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.AbortWithStatusJSON(http.StatusAccepted, urr)
	}
}

func Serve() {
	app := gin.Default()
	api_user := app.Group("/user")
	{
		api_user.POST("/", addUserHandler())
		api_user.DELETE("/", removeUserHandler())
	}
	log.Fatal(app.Run(":3000"))
}
