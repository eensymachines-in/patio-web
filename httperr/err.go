// Errors compatible with http handlers with logging, dispatch codes and client messaging
//
// Errors need to be translated to http status code appropriately.
// They also detailed logging on the server so as to enable precise diagnostics.
// While errors need to send out client messages that provide convincing messages for the status.
// This package builds all of that into the errors, as a circumference of the existing error interface.
package httperr

import (
	"errors"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

// Constructor functions that take in an internal error and send back the error over HttpErr interface.
// NOTE: SetInternal will return nil incase the internal error is nil
var (
	// Http handler context parameter is missing
	ErrContxParamMissing = func(e error) HttpErr {
		return (&eCtxParamMissing{}).SetInternal(e)
	}
	// Error when unmarshalling, typically from Json
	ErrUnMarshal = func(e error) HttpErr {
		return (&eUnmarshal{}).SetInternal(e)
	}
	// DB query has failed, but not when empty result
	ErrDBQuery = func(e error) HttpErr {
		return (&eDBQuery{}).SetInternal(e)
	}
	// Validation of the inputs, arguments went wrong
	ErrValidation = func(e error) HttpErr {
		return (&eValidation{}).SetInternal(e)
	}
	// Context when bound, error - this is similar to unmarshaling except that its used only when used with ShouldBind()
	ErrBinding = func(e error) HttpErr {
		return (&eBinding{}).SetInternal(e)
	}
	// Error publishing message to AMQP server
	ErrSendRabbit = func(e error) HttpErr {
		return (&eSendRabbit{}).SetInternal(e)
	}
	// Error connecting to gateway - DB or AMQP
	ErrGatewayConnect = func(e error) HttpErr {
		return (&eGtwyConn{}).SetInternal(e)
	}
	// Login / authentication failed
	ErrAuthentication = func(e error) HttpErr {
		return (&eUsrAuth{}).SetInternal(e)
	}
	// Failed to get resource when queried
	ErrResourceNotFound = func(e error) HttpErr {
		return (&eResNotFound{}).SetInternal(e)
	}
	// Invalid token, access revoked
	ErrForbidden = func(e error) HttpErr {
		return (&eForbidden{}).SetInternal(e)
	}
	// duplicate resource creation when disallowed
	DuplicateResourceErr = func(e error) HttpErr {
		return (&eDuplicate{})
	}
	// incoming request parameter is invalid
	ErrInvalidParam = func(e error) HttpErr {
		return (&eInvldParam{})
	}
)

type eInvldParam struct {
	Internal error
}

type eDuplicate struct {
	Internal error
}

type eForbidden struct {
	Internal error
}

type eUsrAuth struct {
	Internal error
}

type eCtxParamMissing struct {
	Internal error
}
type eUnmarshal struct {
	Internal error
}
type eDBQuery struct {
	Internal error
}
type eValidation struct {
	Internal error
}
type eBinding struct {
	Internal error
}
type eSendRabbit struct {
	Internal error
}
type eGtwyConn struct {
	Internal error
}

type eResNotFound struct {
	Internal error
}

func (eip *eInvldParam) ClientErrData() string {
	return "One or more parameters for your request was invalid, check and send again."
}

func (edup *eDuplicate) ClientErrData() string {
	return "Duplicate resource isn't allowed, check if the resource is already registered."
}

func (eu *eForbidden) ClientErrData() string {
	return "Unauthorized to access this data on the website. Either you have no elevation or you require to re-login."
}

func (rnf *eResNotFound) ClientErrData() string {
	return "Resource you were looking for is not found, check and send again"
}

func (eua *eUsrAuth) ClientErrData() string {
	return "Authentication failed, check your password and try again."
}
func (ecpm *eCtxParamMissing) ClientErrData() string {
	return "One or more components of the server isn't working as it should be. Kindly contact the system admin"
}
func (eu *eUnmarshal) ClientErrData() string {
	return "Error reading in the data sent, check your inputs and try again. If the problem persists then contact sys admin"
}
func (edbq *eDBQuery) ClientErrData() string {
	return "Failed to fetch data, this happens when one or more components on the server isnt working as expected. Get in touch with a system admin."
}
func (ev *eValidation) ClientErrData() string {
	return "One or more inputs in the request were invalidated by the server, check your inputs and send again"
}
func (eb *eBinding) ClientErrData() string {
	return "Request payload wasnt as expected, check your inputs and send again"
}
func (esr *eSendRabbit) ClientErrData() string {
	return "Couldn't dispatch the changes downstream. One or more gateways rejected, the server will not attempt this again."
}
func (egc *eGtwyConn) ClientErrData() string {
	return "One or more gateways for the server is down/forbidden/closed, server will not attempt to send this again."
}

func (eip *eInvldParam) Error() string {
	return eip.Internal.Error()
}

func (edup *eDuplicate) Error() string {
	return edup.Internal.Error()
}

func (eu *eForbidden) Error() string {
	return eu.Internal.Error()
}

func (rnf *eResNotFound) Error() string {
	return rnf.Internal.Error()
}

func (eua *eUsrAuth) Error() string {
	return eua.Internal.Error()
}

func (ecpm *eCtxParamMissing) Error() string {
	return ecpm.Internal.Error()
}
func (eu *eUnmarshal) Error() string {
	return eu.Internal.Error()
}
func (edbq *eDBQuery) Error() string {
	return fmt.Sprintf("%s", edbq.Internal)
}
func (ev *eValidation) Error() string {
	return ev.Internal.Error()
}
func (eb *eBinding) Error() string {
	return eb.Internal.Error()
}
func (esr *eSendRabbit) Error() string {
	return esr.Internal.Error()
}

func (egc *eGtwyConn) Error() string {
	return egc.Internal.Error()
}

func (eip *eInvldParam) SetInternal(ie error) HttpErr {
	if ie == nil {
		return nil
	}
	eip.Internal = ie
	return eip
}

func (edup *eDuplicate) SetInternal(ie error) HttpErr {
	if ie == nil {
		return nil
	}
	edup.Internal = ie
	return edup
}

func (eu *eForbidden) SetInternal(ie error) HttpErr {
	if ie == nil {
		return nil
	}
	eu.Internal = ie
	return eu
}

func (rnf *eResNotFound) SetInternal(ie error) HttpErr {
	if ie == nil {
		return nil
	}
	rnf.Internal = ie
	return rnf
}

func (eua *eUsrAuth) SetInternal(ie error) HttpErr {
	if ie == nil {
		return nil
	}
	eua.Internal = ie
	return eua
}

func (egc *eGtwyConn) SetInternal(ie error) HttpErr {
	if ie == nil {
		return nil
	}
	egc.Internal = ie
	return egc
}

func (esr *eSendRabbit) SetInternal(ie error) HttpErr {
	if ie == nil {
		return nil
	}
	esr.Internal = ie
	return esr
}

func (ecpm *eCtxParamMissing) SetInternal(ie error) HttpErr {
	if ie == nil {
		return nil
	}
	ecpm.Internal = ie
	return ecpm
}
func (eu *eUnmarshal) SetInternal(ie error) HttpErr {
	if ie == nil {
		return nil
	}
	eu.Internal = ie
	return eu
}
func (edbq *eDBQuery) SetInternal(ie error) HttpErr {
	if ie == nil {
		return nil
	}
	edbq.Internal = ie
	return edbq
}
func (ev *eValidation) SetInternal(ie error) HttpErr {
	if ie == nil {
		return nil
	}
	ev.Internal = ie
	return ev
}
func (eb *eBinding) SetInternal(ie error) HttpErr {
	if ie == nil {
		return nil
	}
	eb.Internal = ie
	return eb
}

/* ========================
Log implementation for each of the errors
this helps to log errors on their way out dispatch
Each error implements Log function to help interfacing functions to uniformly call Log function
========================*/

func (eip *eInvldParam) Log(le *log.Entry) HttpErr {
	le.WithFields(log.Fields{
		"internal_err": eip.Internal,
	}).Error("invalid request parameter")
	return eip
}

func (edup *eDuplicate) Log(le *log.Entry) HttpErr {
	le.WithFields(log.Fields{
		"internal_err": edup.Internal,
	}).Error("duplicate resource")
	return edup
}

func (eu *eForbidden) Log(le *log.Entry) HttpErr {
	le.WithFields(log.Fields{
		"internal_err": eu.Internal,
	}).Error("unauthorized token")
	return eu
}

func (rnf *eResNotFound) Log(le *log.Entry) HttpErr {
	le.WithFields(log.Fields{
		"internal_err": rnf.Internal,
	}).Error("resource now found")
	return rnf
}

func (eua *eUsrAuth) Log(le *log.Entry) HttpErr {
	le.WithFields(log.Fields{
		"internal_err": eua.Internal,
	}).Error("failed user authentication")
	return eua
}

func (egc *eGtwyConn) Log(le *log.Entry) HttpErr {
	le.WithFields(log.Fields{
		"internal_err": egc.Internal,
	}).Error("failed gateway connection")
	return egc
}
func (esr *eSendRabbit) Log(le *log.Entry) HttpErr {
	le.WithFields(log.Fields{
		"internal_err": esr.Internal,
	}).Error("failed sending to rabbitmq")
	return esr
}
func (ecpm *eCtxParamMissing) Log(le *log.Entry) HttpErr {
	le.WithFields(log.Fields{
		"internal_err": ecpm.Internal,
	}).Error("one or more context params in the handler missing")
	return ecpm
}
func (eu *eUnmarshal) Log(le *log.Entry) HttpErr {
	le.WithFields(log.Fields{
		"internal_err": eu.Internal,
	}).Error("failed unmarshalling")
	return eu
}
func (edbq *eDBQuery) Log(le *log.Entry) HttpErr {
	if errors.Is(edbq.Internal, mongo.ErrNoDocuments) {
		le.WithFields(log.Fields{
			"internal_err": edbq.Internal,
		}).Error("empty result query")
	} else {
		le.WithFields(log.Fields{
			"internal_err": edbq.Internal,
		}).Error("failed query on database")
	}
	return edbq
}
func (ev *eValidation) Log(le *log.Entry) HttpErr {
	le.WithFields(log.Fields{
		"internal_err": ev.Internal,
	}).Error("one or more field validations failed")
	return ev
}
func (eb *eBinding) Log(le *log.Entry) HttpErr {
	le.WithFields(log.Fields{
		"internal_err": eb.Internal,
	}).Error("json binding failed")
	return eb
}

/* ========================
Status code embeddingiin the the context requires status code setting from within each error
EAch error corresponds to a single http status code
Interfacing functions use this to embedd sattus code in the gin.Conetext
========================*/

func (eip *eInvldParam) HttpStatusCode() int {
	return http.StatusBadRequest
}

func (edup *eDuplicate) HttpStatusCode() int {
	return http.StatusBadRequest
}
func (eu *eForbidden) HttpStatusCode() int {
	// https://stackoverflow.com/questions/3297048/403-forbidden-vs-401-unauthorized-http-responses
	// 401 eForbidden is actually Unauthenticated and should be used when logging in
	// While 403 is Forbidden and used once the token is generated.
	return http.StatusForbidden
}
func (rnf *eResNotFound) HttpStatusCode() int {
	return http.StatusNotFound
}

func (eua *eUsrAuth) HttpStatusCode() int {
	return http.StatusUnauthorized
}

func (egc *eGtwyConn) HttpStatusCode() int {
	return http.StatusBadGateway
}

func (esr *eSendRabbit) HttpStatusCode() int {
	return http.StatusBadGateway
}

func (ecpm *eCtxParamMissing) HttpStatusCode() int {
	return http.StatusInternalServerError
}
func (eu *eUnmarshal) HttpStatusCode() int {
	return http.StatusInternalServerError
}
func (edbq *eDBQuery) HttpStatusCode() int {
	if errors.Is(edbq.Internal, mongo.ErrNoDocuments) {
		return http.StatusNotFound
	} else {
		return http.StatusBadGateway
	}
}
func (ev *eValidation) HttpStatusCode() int {
	return http.StatusBadRequest
}
func (eb *eBinding) HttpStatusCode() int {
	return http.StatusBadRequest
}
