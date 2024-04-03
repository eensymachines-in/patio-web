package httperr_test

import (
	"fmt"
	"net/http"

	"github.com/eensymachines-in/patio-web/httperr"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Example() {
	// this could be any operation on the server, typically a database operation
	operation := func() httperr.HttpErr { // notice how the operation returns not error, but HttpErr
		// BindingErr is just for instance
		return httperr.ErrBinding(fmt.Errorf("this is an internal error"))
	}
	handler := func(c *gin.Context) { // this is your gin handler
		if err := operation(); err != nil {
			httperr.HttpErrOrOkDispatch(c, err, logrus.WithFields(logrus.Fields{
				"stack": "NameOfTheHandler/Method",
			}))
			return
		}
		c.AbortWithStatus(http.StatusOK) // no error this will dispatch the payload
	}
	// Output:
	// There isnt any real output to this
}
