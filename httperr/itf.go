package httperr

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type HttpErr interface {
	SetInternal(ie error) HttpErr // sets the internal error object, used in logging has internal server information, never send on ClientErrData
	Log(le *log.Entry) HttpErr    // logs the error all in 1 place
	HttpStatusCode() int          // status code relevant to the error
	ClientErrData() string        // error message dispatched to the client (web) typically used with AbortStatusWithJSON
}

// HttpErrOrOkDispatch: given the gin Context, error, and the logging fields this can call c.AbortStatusWithJson, attach appropriate http status code and log the error
func HttpErrOrOkDispatch(c *gin.Context, err HttpErr, le *log.Entry) {
	if err == nil {
		return
	}
	c.AbortWithStatusJSON(err.Log(le).HttpStatusCode(), gin.H{
		"err_data": err.ClientErrData(),
	})
}
