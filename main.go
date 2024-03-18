package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	webpages      = "web/*.html"
	testDeviceUid = "705b200f-059b-4630-bf3d-5d55c3a4a9dc" // this is when we havent got the device registered on the server
)

var (
	db_USER   string
	db_PASSWD string
	db_SERVER string = "srv_mongo"
)

func init() { // logging setup
	log.SetFormatter(&log.TextFormatter{DisableColors: false, FullTimestamp: false})
	log.SetReportCaller(false)

	val := os.Getenv("FLOG")
	if val == "1" {
		f, err := os.Open(os.Getenv("LOGF")) // file for logs
		if err != nil {
			log.SetOutput(os.Stdout) // error in opening log file
			log.Debug("log output is Stdout")
		}
		log.SetOutput(f) // log output set to file direction
		log.Debug("log output is set to file")

	} else {
		log.SetOutput(os.Stdout)
		log.Debug("log output is Stdout")
	}
	val = os.Getenv("SILENT")
	if val == "1" {
		log.SetLevel(log.ErrorLevel) // for development
	} else {
		log.SetLevel(log.DebugLevel) // for production
	}
	fromSecretFile := func(path string) (string, error) {
		f, err := os.Open(path)
		if err != nil {
			return "", fmt.Errorf("error opening the secrets file")
		}
		byt, err := io.ReadAll(f)
		if err != nil {
			return "", fmt.Errorf("error reading the secrets file")
		}
		return string(byt), nil
	}
	// Getting db credentials
	data := map[string]string{
		"userid": "/run/secrets/db_root_username",
		"passwd": "/run/secrets/db_root_password",
	}
	var err error
	db_USER, err = fromSecretFile(data["userid"])
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to load database credentials from secrets file")
	} // loaded user for database login
	db_PASSWD, err = fromSecretFile(data["passwd"])
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to load database credentials from secrets file")
	} // loaded pw for database login
}

func main() {
	log.Info("Now starting the patio-web program..")
	defer log.Warn("Now closing the patio-web program...")

	gin.SetMode(gin.DebugMode)
	r := gin.Default()
	r.LoadHTMLGlob(webpages) // for static content , define all the web content delivery only after this
	r.Static("/assets", "./web/assets/")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
	r.GET("/settings", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
	api := r.Group("/api")
	api.Use(CORS)

	api.POST("/login", func(c *gin.Context) {
		login := Login{}
		byt, err := io.ReadAll(c.Request.Body)
		if err != nil { //ill formed payload
			c.JSON(http.StatusBadRequest, gin.H{})
			return
		}
		defer c.Request.Body.Close()
		if err := json.Unmarshal(byt, &login); err != nil { // unexpected format of the payload
			c.JSON(http.StatusBadRequest, gin.H{})
			return
		}
		if login.UserId != "niranjan_awati" && login.UserId != "tejas_cholkar" {
			c.JSON(http.StatusUnauthorized, gin.H{})
			return
		}
		if (login.UserId == "niranjan_awati" && login.Password == "280382") || (login.UserId == "tejas_cholkar" && login.Password == "040981") {
			c.JSON(http.StatusOK, gin.H{})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{})
		}
	})

	// ------------ CRUD device configurations -------
	devices := api.Group("/devices")
	devices.GET("/:uid/config", MongoConnect, HndlDeviceConfig)
	devices.PUT("/:uid/config", MongoConnect, HndlDeviceConfig)

	log.Fatal(r.Run(":8080"))
}
