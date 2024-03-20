package httperr

import (
	"errors"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrContxParamMissing = func(e error) HttpErr {
		return (&eCtxParamMissing{}).SetInternal(e)
	}
	ErrUnMarshal = func(e error) HttpErr {
		return (&eUnmarshal{}).SetInternal(e)
	}
	ErrDBQuery = func(e error) HttpErr {
		return (&eDBQuery{}).SetInternal(e)
	}
	ErrValidation = func(e error) HttpErr {
		return (&eValidation{}).SetInternal(e)
	}
	ErrBinding = func(e error) HttpErr {
		return (&eBinding{}).SetInternal(e)
	}
	ErrSendRabbit = func(e error) HttpErr {
		return (&eSendRabbit{}).SetInternal(e)
	}
	ErrGatewayConnect = func(e error) HttpErr {
		return (&eGtwyConn{}).SetInternal(e)
	}
)

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

func (egc *eGtwyConn) Log(le *log.Entry) HttpErr {
	le.WithFields(log.Fields{
		"internal_err": egc.Internal,
	}).Error("failed sending message to rabbitmq")
	return egc
}
func (esr *eSendRabbit) Log(le *log.Entry) HttpErr {
	le.WithFields(log.Fields{
		"internal_err": esr.Internal,
	}).Error("failed sending message to rabbitmq")
	return esr
}
func (ecpm *eCtxParamMissing) Log(le *log.Entry) HttpErr {
	le.WithFields(log.Fields{
		"internal_err": ecpm.Internal,
	}).Error("One or more context params is missing, check the handlers the sequence of middleware")
	return ecpm
}
func (eu *eUnmarshal) Log(le *log.Entry) HttpErr {
	le.WithFields(log.Fields{
		"internal_err": eu.Internal,
	}).Error("failed unmarshalling error")
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
	}).Error("one or more field validations failed")
	return eb
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
