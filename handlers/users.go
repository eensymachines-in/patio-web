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
	"github.com/eensymachines-in/patio-web/devices"
	"github.com/eensymachines-in/patio-web/httperr"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

// HndlDevices : handling device / collection of devices  - devices when registered, unregistered, or simply queried for registrations
// Devices store information of the platform and configuration of the actual devices on the ground.
// incase of errors this dispatch as requried after having logged the error
// Needs mongo connection to appropriate "devices" collection

func HndlDevices(c *gin.Context) {
	val, _ := c.Get("mongo-client") // unless you havent got MongoConnect inline middleware this will not require error handling
	mongoClient := val.(*mongo.Client)
	val, _ = c.Get("mongo-database")
	db := val.(*mongo.Database)
	dc := devices.DevicesCollection{DbColl: db.Collection("devices")}
	defer mongoClient.Disconnect(context.Background())

	if c.Request.Method == "POST" { // when the device on the ground seeks register itself
		device := devices.Device{} // resultant device, or the device being inserted
		if err := c.ShouldBind(&device); err != nil {
			httperr.HttpErrOrOkDispatch(c, httperr.ErrBinding(fmt.Errorf("failed to read device from payload")), log.WithFields(log.Fields{
				"stack": "HndlDevices",
			}))
			return
		}
		err := dc.AddDevice(&device) // adds a new device to the registered device
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
				"stack": "HndlDevices",
			}))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, device)
	} else if c.Request.Method == "DELETE" { // device on the ground seeks to unregister itself
		_, err := devices.EitherMacIDOrObjID(c.Param("uid"), dc.RemoveDevice)
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
				"stack": "HndlDevices",
			}))
			return
		}
		c.AbortWithStatus(http.StatusOK)
	} else if c.Request.Method == "GET" { // gets the device
		devc, err := devices.EitherMacIDOrObjID(c.Param("uid"), dc.GetDevice) // typically when the device on the ground needs to check for its registration
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
				"stack": "HndlDevices",
			}))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, devc)
	}
}

// HndlUserDevices : Every devices has legal users, which are enlisted on device.Users
// When logged in User would like to see all the devices under his control
// then change the configuration or see the data accumlated.
// Device when registering itself shall send the users, but incase some users need to be added ahead of the registration it would be using the patch method here
// data on the server shall remain the only single source of truth
func HndlUserDevices(c *gin.Context) {
	val, _ := c.Get("mongo-client") // unless you havent got MongoConnect inline middleware this will not require error handling
	mongoClient := val.(*mongo.Client)
	val, _ = c.Get("mongo-database")
	db := val.(*mongo.Database)
	dc := devices.DevicesCollection{DbColl: db.Collection("devices")}
	defer mongoClient.Disconnect(context.Background())

	if c.Request.Method == "GET" {
		devices, err := dc.UserDevices(c.Param("id")) // object id as string for the user
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
				"stack": "HndlUserDevices",
			}))
			return
		}
		// we have received devices
		c.AbortWithStatusJSON(http.StatusOK, devices)
	} else if c.Request.Method == "PATCH" {
		payload := struct {
			UsersToAdd []string `json:"users"`
		}{}
		if err := c.ShouldBind(&payload); err != nil {
			httperr.HttpErrOrOkDispatch(c, httperr.ErrBinding(err), log.WithFields(log.Fields{
				"stack": "HndlUserDevices/PATCH",
			}))
			return
		}
		updated, err := devices.EitherMacIDOrObjID(c.Param("uid"), dc.UpdateDevice(payload.UsersToAdd)) // for appending user email to the device datbase.
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
				"stack": "HndlUserDevices/PATCH",
			}))
			return
		}
		// we have received devices
		c.AbortWithStatusJSON(http.StatusOK, updated)
	}

}

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
	val, _ := c.Get("mongo-client")
	mongoClient := val.(*mongo.Client)
	val, _ = c.Get("mongo-database")
	db := val.(*mongo.Database)
	uc := auth.UsersCollection{DbColl: db.Collection("users")}
	defer mongoClient.Disconnect(context.Background())

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
