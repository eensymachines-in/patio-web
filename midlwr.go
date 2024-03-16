package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoConnect : will connect to a mongo database and insert the connection to downstream handlers
func MongoConnect(c *gin.Context) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	// credential := options.Credential{
	// 	AuthMechanism: "PLAIN",
	// 	Username:      db_USER,
	// 	Password:      db_PASSWD,
	// }
	// defer cancel()
	// client, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:27017", db_SERVER)).SetAuth(credential))
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s:27017", db_USER, db_PASSWD, db_SERVER)))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Error("failed database conenction: MongoConnect")
		c.JSON(http.StatusBadGateway, gin.H{
			"data": "we are facing connectivity problems for now, try again later",
		})
	}
	if client.Ping(ctx, readpref.Primary()) != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Error("failed database ping: MongoConnect")
		c.JSON(http.StatusBadGateway, gin.H{
			"data": "we are facing issues with data connection, try back again",
		})
		return
	}
	logrus.Debug("connected to database..")
	c.Set("mongo-client", client)
}

// CORS : this allows all cross origin requests
func CORS(c *gin.Context) {
	// First, we add the headers with need to enable CORS
	// Make sure to adjust these headers to your needs
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "*")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Content-Type", "application/json")
	// Second, we handle the OPTIONS problem
	if c.Request.Method != "OPTIONS" {
		c.Next()
	} else {
		// Everytime we receive an OPTIONS request,
		// we just return an HTTP 200 Status Code
		// Like this, Angular can now do the real
		// request using any other method than OPTIONS
		c.AbortWithStatus(http.StatusOK)
	}
}
