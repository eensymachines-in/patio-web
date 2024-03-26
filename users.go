package main

import (
	"net/http"

	"github.com/eensymachines-in/patio-web/auth"
	"github.com/eensymachines-in/patio-web/httperr"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

func HndlUsers(c *gin.Context) {
	usr := auth.User{}
	err := c.ShouldBind(&usr)
	if err != nil {
		httperr.HttpErrOrOkDispatch(c, httperr.ErrBinding(err), log.WithFields(log.Fields{
			"stack": "HndlUserAuth",
		}))
		return
	}
	val, _ := c.Get("mongo-client")
	mongoClient := val.(*mongo.Client)
	uc := auth.UsersCollection{DbColl: mongoClient.Database("aquaponics").Collection("users")}
	if c.Request.Method == "POST" { // login requests are posts requests
		err := uc.NewUser(&usr)
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
				"stack": "HndlUserAuth",
			}))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, usr)
	}
}
func HndlUserAuth(c *gin.Context) {
	usr := auth.User{}
	err := c.ShouldBind(&usr)
	if err != nil {
		httperr.HttpErrOrOkDispatch(c, httperr.ErrBinding(err), log.WithFields(log.Fields{
			"stack": "HndlUserAuth",
		}))
		return
	}
	// getting the mongo connection
	val, _ := c.Get("mongo-client")
	mongoClient := val.(*mongo.Client)
	uc := auth.UsersCollection{DbColl: mongoClient.Database("aquaponics").Collection("users")}
	if c.Request.Method == "POST" { // login requests are posts requests
		err := uc.Authenticate(&usr)
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
				"stack": "HndlUserAuth",
			}))
			return
		}
		// time to send back the token
		c.AbortWithStatusJSON(http.StatusOK, usr)
	}
}
