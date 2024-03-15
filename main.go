package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	webpages = "web/*.html"
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

	api := r.Group("/api").Use(CORS)
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

	log.Fatal(r.Run(":8080"))
}
