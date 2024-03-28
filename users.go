package main

/* =========================
project 		: ipatio-web
date			: MArch` 2024
author			: kneerunjun@gmail.com
Copyrights		: Eensy Machines
About			: Handlers for USER CRUD & authentication when using the GIN framework
				: USes the httperr Error framework for uniform error handling in requests
============================*/
import (
	"fmt"
	"net/http"

	"github.com/eensymachines-in/patio-web/auth"
	"github.com/eensymachines-in/patio-web/httperr"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

// HndlUsers : Handle all routes that require User CRUD
//
// users are uniquely identified by email
//
// uses the mongo connection
//
// POST/ PUT require user details - bound to auth.User{}
//
// GET requires only email
func HndlUsers(c *gin.Context) {
	// Bindings ..
	usr := auth.User{}
	err := httperr.ErrBinding(c.ShouldBind(&usr))
	if err != nil {
		httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
			"stack": "HndlUsers",
		})) // binding failed -
		return
	}
	// Mongo connections
	val, _ := c.Get("mongo-client") // unless you havent got MongoConnect inline middleware this will not require error handling
	mongoClient := val.(*mongo.Client)
	uc := auth.UsersCollection{DbColl: mongoClient.Database("aquaponics").Collection("users")}

	// Actual handler and response
	if c.Request.Method == "POST" { // login requests are posts requests
		err := uc.NewUser(&usr)
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
				"stack": "HndlUsers",
			}))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, &usr)
	} else if c.Request.Method == "DELETE" {
		err := uc.DeleteUser(c.Param("email"))
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
				"stack": "HndlUsers",
			}))
			return
		}
		c.AbortWithStatus(http.StatusOK)
	} else if c.Request.Method == "PATCH" {
		// the identifier shall be used from the url param and not from the payload body
		// identifier from url can be email or id hex  - EditUser has to figure it out
		err := uc.EditUser(c.Param("email"), usr.Name, usr.Auth, usr.TelegID)
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
				"stack": "HndlUsers",
			}))
			return
		}
		c.AbortWithStatus(http.StatusOK)
	}
}

// HndlUserAuth : Handles only user authentication, POST = login, GET = authorization
//
// # JWTtoken for authorization and thru authentication
//
// 401 when password fails
func HndlUserAuth(c *gin.Context) {
	// --------- request biningg
	usr := auth.User{}
	err := httperr.ErrBinding(c.ShouldBind(&usr))
	if err != nil {
		httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
			"stack": "HndlUserAuth",
		}))
		return
	}
	// --------- mongo connections
	val, _ := c.Get("mongo-client")
	mongoClient := val.(*mongo.Client)
	uc := auth.UsersCollection{DbColl: mongoClient.Database("aquaponics").Collection("users")}

	// --------- request handling
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
	} else if c.Request.Method == "GET" {
		// Trying to authorize the requests
		tok := c.Request.Header.Get("Authorization")
		if tok == "" {
			httperr.HttpErrOrOkDispatch(c, httperr.ErrForbidden(fmt.Errorf("empty token cannot request authorization")), log.WithFields(log.Fields{
				"stack": "HndlUserAuth",
			}))
			return
		} else {
			err := uc.Authorize(tok) // user fields would be empty per say since its only the token you are authorizing
			if err != nil {
				httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
					"stack": "HndlUserAuth",
				}))
				return
			}
			c.AbortWithStatus(http.StatusOK)
		}
	}
}
