package main

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
	"time"

	"github.com/eensymachines-in/patio-web/httperr"
	"github.com/eensymachines-in/patio/aquacfg"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
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

func HndlDeviceConfig(c *gin.Context) {
	val, exists := c.Get("mongo-client")
	if !exists {
		httperr.HttpErrOrOkDispatch(c, httperr.ErrContxParamMissing(fmt.Errorf("missing mongo-client")), log.WithFields(log.Fields{
			"stack": "HndlDeviceConfig",
		}))
		return
	}
	/* Database connections*/
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, _ := val.(*mongo.Client)
	defer client.Disconnect(ctx)
	db := client.Database("devices")
	configs := db.Collection("configs")

	// Getting the device uid
	uid := c.Param("uid")
	// uid = testDeviceUid // TODO: remove this line to bypass hardcoding

	if c.Request.Method == "GET" { //getting the device configuration from the server
		// TODO:  replace testDeviceUID from the url params as the actula device uid
		res := configs.FindOne(ctx, bson.M{"uid": uid}) // getting device by its uniq id, generally the mac id
		if res.Err() != nil {
			httperr.HttpErrOrOkDispatch(c, httperr.ErrDBQuery(res.Err()), log.WithFields(log.Fields{
				"stack": "HndlDeviceConfig/GET",
				"uid":   uid,
			}))
			return
		}
		ac := aquacfg.AppConfig{}
		err := res.Decode(&ac.Schedule)
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, httperr.ErrUnMarshal(err), log.WithFields(log.Fields{
				"stack":    "HndlDeviceConfig/GET",
				"schedule": fmt.Sprintf("%v", ac.Schedule),
			}))
			return
		}
		c.JSON(http.StatusOK, &ac.Schedule)
		return
	} else if c.Request.Method == "PUT" { // webapp woudl want to change the device configuration
		ac := aquacfg.AppConfig{}
		err := c.ShouldBind(&ac.Schedule)
		if err != nil {
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
		// Incase the database update succeeds but messageq exchange isnt reachable that would put server out of sync from the device below
		// And since the messageq hasnt received the message there is no way for the device to pull up the message
		backup := aquacfg.AppConfig{}
		err = configs.FindOne(ctx, bson.M{"uid": uid}).Decode(&backup.Schedule) // getting the backup of configuration incase mq does fails.
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, httperr.ErrDBQuery(err), log.WithFields(log.Fields{
				"stack": "HndlDeviceConfig/PUT",
				"uid":   uid,
			}))
			return
		}

		_, err = configs.UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": ac.Schedule})
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, httperr.ErrDBQuery(err), log.WithFields(log.Fields{
				"stack": "HndlDeviceConfig/PUT",
			}))
			return
		}
		byt, _ := json.Marshal(ac.Schedule)
		// Getting the rabbitmq connection so as to set up flushing of it before handler exits
		val, exists = c.Get("rabbit-conn")
		if !exists {
			httperr.HttpErrOrOkDispatch(c, httperr.ErrContxParamMissing(fmt.Errorf("missing mongo-client")), log.WithFields(log.Fields{
				"stack": "HndlDeviceConfig/PUT",
			}))
			return
		}
		conn := val.(*amqp.Connection)
		defer conn.Close()

		if err := publishToRabbit(c, "", byt); err != nil {
			configs.UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": backup.Schedule}) // undoing the database changes
			httperr.HttpErrOrOkDispatch(c, httperr.ErrSendRabbit(err), log.WithFields(log.Fields{
				"stack": "HndlDeviceConfig/PUT",
			}))
			return
		}
		c.JSON(http.StatusOK, gin.H{})
		return
	}
}
