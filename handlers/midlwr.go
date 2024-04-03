package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/eensymachines-in/patio-web/auth"
	"github.com/eensymachines-in/patio-web/httperr"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// SingleUserOfID : from :userid will hit the database with users and get the user object, if not found send back a response from here
// Do not use this as the last handler in the chain of handlers since this does not close the database connections
// Use only after MongoConnect middleware since this uses the mongo connection
func SingleUserOfID(c *gin.Context) {
	// val, _ := c.Get("mongo-client")
	// mongoClient := val.(*mongo.Client)
	// No use for client, uses the database directly
	val, _ := c.Get("mongo-database")
	db := val.(*mongo.Database)
	uc := auth.UsersCollection{DbColl: db.Collection("users")}
	user := auth.User{}
	err := uc.FindUser(c.Param("userid"), &user)
	if err != nil {
		httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
			"stack": "SingleUserOfID",
		})) // binding failed -
		return
	}
	// If user is found this shall set it in context and call next
	c.Set("user", user)
	c.Next()
}

func Authorize(c *gin.Context) {
	val, _ := c.Get("mongo-client")
	mongoClient := val.(*mongo.Client)
	uc := auth.UsersCollection{DbColl: mongoClient.Database("aquaponics").Collection("users")}

	tok := c.Request.Header.Get("Authorization")
	if tok == "" {
		httperr.HttpErrOrOkDispatch(c, httperr.ErrForbidden(fmt.Errorf("empty token, cannot authorize")), log.WithFields(log.Fields{
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
		c.Next() // request is authorized
	}
}

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

func MongoConnect(server, user, passwd, dbname string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s:27017", user, passwd, server)))
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, httperr.ErrGatewayConnect(err), log.WithFields(log.Fields{
				"stack":  "MongoConnect",
				"login":  user,
				"server": server,
			}))
			return
		}
		if client.Ping(ctx, readpref.Primary()) != nil {
			httperr.HttpErrOrOkDispatch(c, httperr.ErrGatewayConnect(err), log.WithFields(log.Fields{
				"stack":  "MongoConnect",
				"login":  user,
				"server": server,
			}))
			return
		}
		c.Set("mongo-client", client)
		c.Set("mongo-database", client.Database(dbname))
	}
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
