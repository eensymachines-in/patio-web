package main

/* =====================
All the middlewaer thats neede to handle / CRUD devices is here
author 		: kneerunjun@gmail.com
project 	: eensymachines aquaponics
place 		: pune
======================*/

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/eensymachines-in/patio/aquacfg"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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

func HndlDeviceConfig(c *gin.Context) {
	val, exists := c.Get("mongo-client")
	if !exists {
		log.Error("Missing database connection from previous handler.. HAve you called MongoConnect before this handler?")
		c.JSON(http.StatusInternalServerError, gin.H{
			"data": "we are facing connectivity problems for now, try again later",
		})
	}
	/* Database connections*/
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, _ := val.(*mongo.Client)
	defer client.Disconnect(ctx)
	db := client.Database("devices")
	configs := db.Collection("configs")

	// Getting the device uid
	uid := c.Param("uid")
	uid = testDeviceUid // TODO: remove this line to bypass hardcoding

	if c.Request.Method == "GET" { //getting the device configuration from the server
		// TODO:  replace testDeviceUID from the url params as the actula device uid
		res := configs.FindOne(ctx, bson.M{"uid": uid}) // getting device by its uniq id, generally the mac id
		if res.Err() != nil {
			if errors.Is(res.Err(), mongo.ErrNoDocuments) {
				log.WithFields(log.Fields{
					"err": res.Err(),
					"uid": uid,
				}).Error("error getting devices: HndlDeviceConfig")
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
					"msg": fmt.Sprintf("Missing device with id: %s", uid),
				})
			} else {
				log.WithFields(log.Fields{
					"err": res.Err(),
					"uid": uid,
				}).Error("error getting devices: HndlDeviceConfig")
				c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{
					"msg": fmt.Sprintf("Error getting device details: %s", uid),
				})
			}
			return
		}
		ac := aquacfg.AppConfig{}
		if err := res.Decode(&ac.Schedule); err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("while unmarshaling data from database: HndlDeviceConfig")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"msg": fmt.Sprintf("Internal error getting device details: %s", uid),
			})
			return
		}
		c.JSON(http.StatusOK, &ac.Schedule)
		return
	} else if c.Request.Method == "PUT" { // webapp woudl want to change the device configuration
		ac := aquacfg.AppConfig{}
		if err := c.ShouldBind(&ac.Schedule); err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to read new config data : HndlDeviceConfig")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"msg": "Error reading request payload:",
			})
			return
		}
		if !validateDeviceSched(ac) {
			log.Error("Invalid schedule for the device : HndlDeviceConfig")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"msg": "Invalid device configuration, check and send again",
			})
			return
		}
		res, err := configs.UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": ac.Schedule})
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				log.WithFields(log.Fields{
					"matched": res.MatchedCount,  // shoudl be 0
					"updated": res.UpsertedCount, // should be 0
				}).Error("no matching devices found to update configuration: HndlDeviceConfig")
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
					"msg": fmt.Sprintf("Error updating device configuration for %s, no device found", uid),
				})
			} else {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("failed to update device configuration: HndlDeviceConfig")
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"msg": "Internal error updating device configuration",
				})
				return
			}
		}
		// TODO: here we need to post an update message to message queue so as to update the device accordingly
		// NOTE: the source of truth will always be this go server - the client updates the configuration and push notifications are for the device which follows this
		c.JSON(http.StatusOK, gin.H{})
		return
	}
}
