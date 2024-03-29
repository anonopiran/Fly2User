package api2user

import (
	config "Fly2User/Config"
	supervisor "Fly2User/Supervisor"
	"errors"
	"net/http"

	u2u "Fly2User/User2Upstream"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func manUser(fn func(*grpc.ClientConn, *[]u2u.InboundType, bool) []u2u.GrpcError, exitOnErr bool) bool {
	hasErr := false
	for _, trg := range cfg().V2flyApiAddress {
		logWithTarget := logrus.WithField("target", trg)
		res, err := u2u.ResolveV2FlyServer(trg.AsUrl())
		if err != nil {
			logWithTarget.WithError(err).Error("error while resolving v2fly server")
			continue
		}
		for _, srv := range res {
			logWithServerData := logrus.WithField("server", srv)
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
			iList := cfg().InboundList
			errList := fn(conn, &iList, false)
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
	}
	return hasErr
}
func addUser(user *u2u.UserAddType, exitOnErr bool) bool {
	return manUser(user.AddMultiple, exitOnErr)
}
func rmUser(user *u2u.UserRemoveType, exitOnErr bool) bool {
	return manUser(user.RemoveMultiple, exitOnErr)
}
func addUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logWithField := logrus.WithField("handler", "removeUserHandler")
		uar := u2u.UserAddType{}
		if err := c.BindJSON(&uar); err != nil {
			logWithField.WithError(err).Error()
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		email := uar.Email
		if supervisor.UserCheckFile(email) {
			existing, err := supervisor.UserFromFile(email)
			if err != nil {
				logWithField.WithError(err).Error()
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			req_hash, err := hashstructure.Hash(uar, hashstructure.FormatV2, nil)
			if err != nil {
				logWithField.WithError(err).Error()
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			existing_hash, err := hashstructure.Hash(existing, hashstructure.FormatV2, nil)
			if err != nil {
				logWithField.WithError(err).Error()
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			if req_hash != existing_hash {
				c.AbortWithStatusJSON(http.StatusConflict, map[string]string{"msg": "user with same email and different config already exists"})
				return
			}
		}

		// ........
		urr := u2u.UserRemoveType{Email: uar.Email}
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
		logWithField := logrus.WithField("handler", "removeUserHandler")
		urr := u2u.UserRemoveType{}
		if err := c.BindJSON(&urr); err != nil {
			logWithField.WithError(err).Error()
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
			logWithField.WithError(err).Error()
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.AbortWithStatusJSON(http.StatusAccepted, urr)
	}
}
func countUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logWithField := logrus.WithField("handler", "countUserHandler")
		usrs, err := supervisor.ReadAllUsers()
		if err != nil {
			logWithField.WithError(err).Error()
			c.AbortWithError(http.StatusInternalServerError, errors.New("an unknown error happened while calling upstream server"))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, len(usrs))
	}
}
func cfg() config.SettingsType {
	return *config.Config()
}
