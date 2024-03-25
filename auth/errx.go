package auth

import "fmt"

var (
	InvalidUserErr = func(e error) error {
		return &InvalidUser{Internal: e}
	}
	FailedDBQueryErr = func(e error) error {
		return &DBQueryFail{Internal: e}
	}
	DuplicateEntryErr = func(e error) error {
		return &DuplicateEntry{}
	}
	UserNotFoundErr = func(e error) error {
		return &UserNotFound{}
	}
	MismatchPasswdErr = func(e error) error {
		return &MismatchPasswd{}
	}
	AuthTokenErr = func(e error) error {
		return &GenTokenFail{}
	}
)

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
