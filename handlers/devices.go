package handlers

/* =====================
All the middlewaer thats neede to handle / CRUD devices is here
author 		: kneerunjun@gmail.com
project 	: eensymachines aquaponics
place 		: pune
======================*/

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eensymachines-in/patio-web/auth"
	"github.com/eensymachines-in/patio-web/devices"
	"github.com/eensymachines-in/patio-web/httperr"
	"github.com/eensymachines-in/patio/aquacfg"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
)

// validateDeviceSched : for all schedule configurations that are incoming it needs a check before its saved on the database
// schedule configurations on the database are single source of truths and hence low tolerance for any rogue values
func validateDeviceSched(cfg aquacfg.AppConfig) bool {
	if cfg.Schedule.Config < aquacfg.ScheduleType(0) || cfg.Schedule.Config > aquacfg.ScheduleType(3) {
		return false
	}
	if cfg.Schedule.Config == aquacfg.PULSE_EVERY { // for other schedule types checking is irrelevant
		// For PULSE_EVERY_DAYAT - interval is irrelevant
		// For TICK_EVERY - pulse gap is irrelevant
		if cfg.Schedule.Interval <= cfg.Schedule.PulseGap {
			return false // interval cannot be equal or less than pulse gap
		}
	}
	if cfg.Schedule.Config == aquacfg.PULSE_EVERY_DAYAT || cfg.Schedule.Config == aquacfg.TICK_EVERY_DAYAT {
		if cfg.Schedule.TickAt == "" { //when clock driven the clock cannot be empty
			return false
		}
	}
	return true
}

// publishToRabbit : knows how to get rabbit pieces from middleware and publish the desired message to intended exchange
// c		: gin context which is laden by middleware pieces
// exchng 	: name of the exchange to publish to
// byt		: byte message to post
// returns error to indicate that handler can exit, loads the context with appropriate error. From the client check for error and if error just return
func publishToRabbit(c *gin.Context, exchng string, byt []byte) error {
	val, exists := c.Get("rabbit-channel")
	if !exists {
		return fmt.Errorf("publishToRabbit: missing context param - rabbit-channel")
	}
	ch := val.(*amqp.Channel)

	val, exists = c.Get("rabbit-queue")
	if !exists {
		return fmt.Errorf("publishToRabbit: missing context param - rabbit-queue")
	}
	q := val.(amqp.Queue) // NOTE: this isnt *amqp.Queue

	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()
	err := ch.Publish(
		exchng, // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        byt,
		},
	)
	if err != nil {
		return fmt.Errorf("failed publishToRabbit: %s", err)
	}
	return nil
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
	// From SingleUserOfID - the user object from the object id is derived and injected in the context
	// user object then can be accessed for its fields
	// example: device to user mapping is always thru email, since user object in the database can be transient but if after deletion the user chooses the same email address then it would automatically connect thee user and the device
	val, _ = c.Get("user")
	user := val.(auth.User)

	if c.Request.Method == "GET" {
		devices, err := dc.UserDevices(user.Email) // object id as string for the user
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
		device := devices.Device{Config: aquacfg.AppConfig{}} // resultant device, or the device being inserted
		if err := c.ShouldBind(&device); err != nil {
			httperr.HttpErrOrOkDispatch(c, httperr.ErrBinding(fmt.Errorf("failed to read device from payload")), log.WithFields(log.Fields{
				"stack": "HndlDevices",
			}))
			return
		}
		log.WithFields(log.Fields{
			"mac":      device.Mac,
			"name":     device.Name,
			"location": device.Location,
			"make":     device.Make,
		}).Debug("new device to add")
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

// HndlDeviceConfig : getting and patching the device configurations
// While the front end shall GET request the device configuration, the once the user submits the configuration PATCHed configuration then shall be communicated to the device over AMQP rabbit
func HndlDeviceConfig(c *gin.Context) {
	val, _ := c.Get("mongo-client") // unless you havent got MongoConnect inline middleware this will not require error handling
	mongoClient := val.(*mongo.Client)
	val, _ = c.Get("mongo-database")
	db := val.(*mongo.Database)
	dc := devices.DevicesCollection{DbColl: db.Collection("devices")}
	defer mongoClient.Disconnect(context.Background())

	if c.Request.Method == "GET" {
		devc, err := devices.EitherMacIDOrObjID(c.Param("uid"), dc.GetDevice) // typically when the device on the ground needs to check for its registration
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
				"stack": "HndlDevices",
			}))
			return
		}
		// NOTE: we send only the schedule of the device and not the entire configuration
		c.AbortWithStatusJSON(http.StatusOK, devc.Config.Schedule)
	} else if c.Request.Method == "PATCH" {
		ac := aquacfg.AppConfig{}
		if err := c.ShouldBind(&ac.Schedule); err != nil {
			httperr.HttpErrOrOkDispatch(c, httperr.ErrBinding(err), log.WithFields(log.Fields{
				"stack":    "HndlDeviceConfig/PUT",
				"schedule": fmt.Sprintf("%v", ac.Schedule),
			}))
			return
		}
		if !validateDeviceSched(ac) {
			httperr.HttpErrOrOkDispatch(c, httperr.ErrValidation(fmt.Errorf("failed validation of device schedule")), log.WithFields(log.Fields{
				"stack":    "HndlDeviceConfig/PUT",
				"schedule": fmt.Sprintf("%v", ac.Schedule),
			}))
			return
		}
		// Updating the device of its schedule  ..
		backup := aquacfg.AppConfig{}
		devc, err := devices.EitherMacIDOrObjID(c.Param("uid"), dc.UpdateDeviceCfg(&ac, &backup))
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
				"stack": "HndlDevices",
			}))
			return
		}
		byt, _ := json.Marshal(ac.Schedule)
		val, exists := c.Get("rabbit-conn")
		if !exists {
			httperr.HttpErrOrOkDispatch(c, httperr.ErrContxParamMissing(fmt.Errorf("missing mongo-client")), log.WithFields(log.Fields{
				"stack": "HndlDeviceConfig/PUT",
			}))
			return
		}
		conn := val.(*amqp.Connection)
		defer conn.Close()
		// Pushing to rabbit mq over amqp for the device to get noticiations.
		if err := publishToRabbit(c, "", byt); err != nil {
			// case when the server was updated of the configuration but the device could not be contacted  ..
			// in case we revert the changes on server to what it was .. so that source of truth stays the same
			// AMQP and database update are atomic operation
			_, err := devices.EitherMacIDOrObjID(c.Param("uid"), dc.UpdateDeviceCfg(&backup, &aquacfg.AppConfig{}))
			if err != nil {
				httperr.HttpErrOrOkDispatch(c, httperr.ErrDBQuery(fmt.Errorf("failure to revert server changes, this means the server is out of sync from the device permanently %s", err)), log.WithFields(log.Fields{
					"stack": "HndlDeviceConfig/PUT",
				}))
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusOK, devc.Config)
		return
	}
}
