package handlers

/* =========================
project 		: ipatio-web
date			: MArch` 2024
author			: kneerunjun@gmail.com
Copyrights		: Eensy Machines
About			: Handlers for USER CRUD & authentication when using the GIN framework
				: USes the httperr Error framework for uniform error handling in requests
============================*/
import (
	"context"
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

	// Mongo connections
	val, _ := c.Get("mongo-client")
	mongoClient := val.(*mongo.Client)
	val, _ = c.Get("mongo-database")
	db := val.(*mongo.Database)
	uc := auth.UsersCollection{DbColl: db.Collection("users")}
	defer mongoClient.Disconnect(context.Background())

	// Actual handler and response
	if c.Request.Method == "POST" { // login requests are posts requests
		usr := auth.User{}
		err := httperr.ErrBinding(c.ShouldBind(&usr))
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
				"stack": "HndlUsers",
			})) // binding failed -
			return
		}
		usr.Role = auth.EndUser // when creating new user the role will always be EndUser
		err = uc.NewUser(&usr)
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
				"stack": "HndlUsers",
			}))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, &usr)
	} else if c.Request.Method == "DELETE" {
		// for deleting the user email is required, but emails in urls arent a great idea
		// Hence we get the user which is injected by SingleUserOfID
		val, _ := c.Get("user")
		usr := val.(auth.User)
		err := uc.DeleteUser(usr.Email)
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
				"stack": "HndlUsers",
			}))
			return
		}
		c.AbortWithStatus(http.StatusOK)
	} else if c.Request.Method == "PATCH" {
		// EditUser takes emails, while email as user identifier in urls isnt a good idea
		// We thus use SingleUserOfID middleware that can get the user object as required
		val, _ := c.Get("user")
		usr := val.(auth.User)
		patch := struct {
			Name    string `json:"name"`
			Auth    string `json:"auth"`
			TelegID int64  `json:"telegid"`
		}{}
		if err := httperr.ErrBinding(c.ShouldBind(&patch)); err != nil {
			httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
				"stack": "HndlUsers",
			}))
			return
		}
		err := uc.EditUser(usr.Email, patch.Name, patch.Auth, patch.TelegID)
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
	val, _ = c.Get("mongo-database")
	db := val.(*mongo.Database)
	uc := auth.UsersCollection{DbColl: db.Collection("users")}
	defer mongoClient.Disconnect(context.Background())

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
