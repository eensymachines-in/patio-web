package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/eensymachines-in/patio-web/httperr"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// RabbitConnectWithChn : name of the channel this publishes to, via default exchange
func RabbitConnectWithChn(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := amqp.Dial(fmt.Sprintf("amqp://%s@%s/", os.Getenv("AMQP_LOGIN"), os.Getenv("AMQP_SERVER")))
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, httperr.ErrGatewayConnect(err), log.WithFields(log.Fields{
				"stack":  "RabbitConnectWithChn",
				"login":  os.Getenv("AMQP_LOGIN"),
				"server": os.Getenv("AMQP_SERVER"),
			}))
			return
		}
		// defer conn.Close()
		ch, err := conn.Channel()
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, httperr.ErrGatewayConnect(err), log.WithFields(log.Fields{
				"stack":  "RabbitConnectWithChn",
				"login":  os.Getenv("AMQP_LOGIN"),
				"server": os.Getenv("AMQP_SERVER"),
			}))
			return
		}
		q, err := ch.QueueDeclare(
			name,  // name
			false, // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, httperr.ErrGatewayConnect(err), log.WithFields(log.Fields{
				"stack":  "RabbitConnectWithChn",
				"login":  os.Getenv("AMQP_LOGIN"),
				"server": os.Getenv("AMQP_SERVER"),
			}))
			return
		}
		c.Set("rabbit-channel", ch) // posting messages for downstream
		c.Set("rabbit-queue", q)    // posting messages for downstream
		c.Set("rabbit-conn", conn)  //since we need to close from downstream
	}
}

// MongoConnect : will connect to a mongo database and insert the connection to downstream handlers
func MongoConnect(c *gin.Context) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s:27017", db_USER, db_PASSWD, db_SERVER)))
	if err != nil {
		httperr.HttpErrOrOkDispatch(c, httperr.ErrGatewayConnect(err), log.WithFields(log.Fields{
			"stack":  "MongoConnect",
			"login":  db_USER,
			"server": db_SERVER,
		}))
		return
	}
	if client.Ping(ctx, readpref.Primary()) != nil {
		httperr.HttpErrOrOkDispatch(c, httperr.ErrGatewayConnect(err), log.WithFields(log.Fields{
			"stack":  "MongoConnect",
			"login":  db_USER,
			"server": db_SERVER,
		}))
		return
	}
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
