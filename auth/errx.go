package auth

/*
All error implementation for auth package,
They implement  the httperr interfaces since eventually they have to be carried as payload in error responses

*/
import (
	"fmt"
	"net/http"

	"github.com/eensymachines-in/patio-web/httperr"
	log "github.com/sirupsen/logrus"
)

var (
	InvalidUserErr = func(e error) httperr.HttpErr {
		return (&InvalidUser{}).SetInternal(e)
	}
	FailedDBQueryErr = func(e error) httperr.HttpErr {
		return (&DBQueryFail{}).SetInternal(e)
	}
	DuplicateEntryErr = func(e error) httperr.HttpErr {
		return (&DuplicateEntry{}).SetInternal(e)
	}
	UserNotFoundErr = func(e error) httperr.HttpErr {
		return (&UserNotFound{}).SetInternal(e)
	}
	MismatchPasswdErr = func(e error) httperr.HttpErr {
		return (&MismatchPasswd{}).SetInternal(e)
	}
	AuthTokenErr = func(e error) httperr.HttpErr {
		return (&GenTokenFail{}).SetInternal(e)
	}
	InvalidTokenErr = func(e error) httperr.HttpErr {
		return (&eInvalidToken{}).SetInternal(e)
	}
)

type eInvalidToken struct {
	Internal error
}

type GenTokenFail struct {
	Internal error
}

type MismatchPasswd struct {
	Internal error
}

type UserNotFound struct {
	Internal error
}

type DBQueryFail struct {
	Internal error
}
type InvalidUser struct {
	Internal error
}

type DuplicateEntry struct {
	Internal error
}

// =================== Implementation of the error interface

func (iu *InvalidUser) Error() string {
	return fmt.Sprintf("Invalid user: %s", iu.Internal)
}
func (dbq *DBQueryFail) Error() string {
	return fmt.Sprintf("Failed DB query: %s", dbq.Internal)
}
func (dbq *DuplicateEntry) Error() string {
	return fmt.Sprintf("Duplicate user: %s", dbq.Internal)
}
func (unf *UserNotFound) Error() string {
	return fmt.Sprintf("Duplicate user: %s", unf.Internal)
}
func (mp *MismatchPasswd) Error() string {
	return fmt.Sprintf("Duplicate user: %s", mp.Internal)
}
func (gt *GenTokenFail) Error() string {
	return fmt.Sprintf("Failed to generate token: %s", gt.Internal)
}

// =================== Impleementation of HttpErr interface

func (it *eInvalidToken) SetInternal(ie error) httperr.HttpErr {
	if ie == nil {
		return nil
	}
	it.Internal = ie
	return it
}

func (iu *InvalidUser) SetInternal(ie error) httperr.HttpErr {
	if ie == nil {
		return nil
	}
	iu.Internal = ie
	return iu
}
func (dbq *DBQueryFail) SetInternal(ie error) httperr.HttpErr {
	if ie == nil {
		return nil
	}
	dbq.Internal = ie
	return dbq
}
func (de *DuplicateEntry) SetInternal(ie error) httperr.HttpErr {
	if ie == nil {
		return nil
	}
	de.Internal = ie
	return de
}
func (unf *UserNotFound) SetInternal(ie error) httperr.HttpErr {
	if ie == nil {
		return nil
	}
	unf.Internal = ie
	return unf
}
func (mp *MismatchPasswd) SetInternal(ie error) httperr.HttpErr {
	if ie == nil {
		return nil
	}
	mp.Internal = ie
	return mp
}
func (gt *GenTokenFail) SetInternal(ie error) httperr.HttpErr {
	if ie == nil {
		return nil
	}
	gt.Internal = ie
	return gt
}

// Logging of the error
func (it *eInvalidToken) Log(le *log.Entry) httperr.HttpErr {
	le.WithFields(log.Fields{
		"internal_err": it.Internal,
	}).Error("invalid or expired token")
	return it
}
func (iu *InvalidUser) Log(le *log.Entry) httperr.HttpErr {
	le.WithFields(log.Fields{
		"internal_err": iu.Internal,
	}).Error("invalid user data..")
	return iu
}
func (dbq *DBQueryFail) Log(le *log.Entry) httperr.HttpErr {
	le.WithFields(log.Fields{
		"internal_err": dbq.Internal,
	}).Error("failed query on DB")
	return dbq
}
func (de *DuplicateEntry) Log(le *log.Entry) httperr.HttpErr {
	le.WithFields(log.Fields{
		"internal_err": de.Internal,
	}).Error("duplicate entry found..")
	return de
}
func (unf *UserNotFound) Log(le *log.Entry) httperr.HttpErr {
	le.WithFields(log.Fields{
		"internal_err": unf.Internal,
	}).Error("user not found registered")
	return unf
}
func (mp *MismatchPasswd) Log(le *log.Entry) httperr.HttpErr {
	le.WithFields(log.Fields{
		"internal_err": mp.Internal,
	}).Error("incorrect password, unauthorized")
	return mp
}
func (gt *GenTokenFail) Log(le *log.Entry) httperr.HttpErr {
	le.WithFields(log.Fields{
		"internal_err": gt.Internal,
	}).Error("failed to generate jwt token")
	return gt
}

// Client error, the one that gets dispatched to the http client
func (it *eInvalidToken) ClientErrData() string {
	return "Your authorization has expired/invalidated, kindly login again"
}
func (iu *InvalidUser) ClientErrData() string {
	return "Invalid user when authenticating it, check all fields and send again"
}
func (dbq *DBQueryFail) ClientErrData() string {
	return "Query failed when authenticating user"
}
func (de *DuplicateEntry) ClientErrData() string {
	return "User can be registered only once, This one is already registered"
}
func (unf *UserNotFound) ClientErrData() string {
	return "User you were trying to authenticate wasnt found"
}
func (mp *MismatchPasswd) ClientErrData() string {
	return "Password did not match our records, authentication failed"
}
func (gt *GenTokenFail) ClientErrData() string {
	return "One or more operations on server side has failed, contact an admin to fix this"
}

// implementing http status code
func (it *eInvalidToken) HttpStatusCode() int {
	return http.StatusForbidden
}
func (iu *InvalidUser) HttpStatusCode() int {
	return http.StatusBadRequest
}
func (dbq *DBQueryFail) HttpStatusCode() int {
	return http.StatusInternalServerError
}
func (de *DuplicateEntry) HttpStatusCode() int {
	return http.StatusBadRequest
}
func (unf *UserNotFound) HttpStatusCode() int {
	return http.StatusNotFound
}
func (mp *MismatchPasswd) HttpStatusCode() int {
	return http.StatusUnauthorized
}
func (gt *GenTokenFail) HttpStatusCode() int {
	return http.StatusInternalServerError
}
