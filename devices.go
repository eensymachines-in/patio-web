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
	"net/http"
	"time"

	"github.com/eensymachines-in/patio/aquacfg"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

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
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{})
			} else {
				log.WithFields(log.Fields{
					"err": res.Err(),
					"uid": uid,
				}).Error("error getting devices: HndlDeviceConfig")
				c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{})
			}
			return
		}
		ac := aquacfg.AppConfig{}
		if err := res.Decode(&ac.Schedule); err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("while unmarshaling data from database: HndlDeviceConfig")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
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
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{})
			return
		}
		res, err := configs.UpdateOne(ctx, bson.M{"uid": testDeviceUid}, bson.M{"$set": ac.Schedule})
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				log.WithFields(log.Fields{
					"matched": res.MatchedCount,  // shoudl be 0
					"updated": res.UpsertedCount, // should be 0
				}).Error("no matching devices found to update configuration: HndlDeviceConfig")
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{})
			} else {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("failed to update device configuration: HndlDeviceConfig")
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
				return
			}
		}
		// TODO: here we need to post an update message to message queue so as to update the device accordingly
		// NOTE: the source of truth will always be this go server - the client updates the configuration and push notifications are for the device which follows this
		c.JSON(http.StatusOK, gin.H{})
		return
	}
}
