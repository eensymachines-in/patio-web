package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	webpages = "web/*.html"
)

func init() { // logging setup
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: false,
		PadLevelText:  true,
	})
	log.SetReportCaller(false)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel) // default is info level, if verbose then trace

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

/*
	serveIndexPage : one place common to send out the index page, base page to the angular app on the client

Angular works as SPA but uses this as the base for further loading all the page content
*/
func serveIndexPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"usrauth_baseurl":   "http://aqua.eensymachines.in:30002/api/users",
		"devicereg_baseurl": "http://aqua.eensymachines.in:30001/api/devices",
	})
}
func main() {
	log.Info("Now starting the patio-web program..")
	defer log.Warn("Now closing the patio-web program...")

	gin.SetMode(gin.DebugMode)
	r := gin.Default()
	r.LoadHTMLGlob(webpages) // for static content , define all the web content delivery only after this
	r.Static("/assets", "./web/assets/")

	r.GET("/", serveIndexPage)
	r.GET("/devices/:deviceID/config", serveIndexPage)
	r.GET("/users/:id/devices", serveIndexPage)
	r.GET("/terms-conditions", serveIndexPage)

	log.Fatal(r.Run(":8080"))
}
