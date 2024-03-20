package httperr

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type HttpErr interface {
	SetInternal(ie error) HttpErr
	Log(le *log.Entry) HttpErr
	HttpStatusCode() int
}

func HttpErrOrOkDispatch(c *gin.Context, err HttpErr, le *log.Entry) {
	if err == nil {
		return
	}
	c.AbortWithStatus(err.Log(le).HttpStatusCode())
}
