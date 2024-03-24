package auth

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
)

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
)

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

type UsersCollection struct {
	DbColl *mongo.Collection
}

type UserPassword string

func (up UserPassword) IsValid() bool {
	passRegex := regexp.MustCompile(`^[0-9a-zA-Z_!@#$%^&*-]{9,12}$`) // password is 9-12 characters
	return passRegex.MatchString(string(up))
}

func (up UserPassword) StringHash() (string, error) {
	h, e := bcrypt.GenerateFromPassword([]byte(string(up)), 14)
	if e != nil {
		return "", e
	}
	return string(h), nil
}

type UserName string

func (un UserName) IsValid() bool {
	passRegex := regexp.MustCompile(`^[a-zA-Z]+$`) // password is 9-12 characters
	return passRegex.MatchString(string(un))
}

type UserEmail string

func (ue UserEmail) IsValid() bool {
	passRegex := regexp.MustCompile(`^[a-zA-Z0-9]+[_.-]*[a-zA-Z0-9]*@[a-zA-Z0-9]+[.]{1}[a-zA-Z0-9]{2,}$`) // password is 9-12 characters
	return passRegex.MatchString(string(ue))
}

// EditUser: this can change the user fields given the email
// can alter name passwd and telegid when editing
// email cannot be altered, create a new account with a new email to use it
func (u *UsersCollection) EditUser(email string, name, passwd string, telegid int64) error {
	ctx, _ := context.WithCancel(context.Background())
	cnt, err := u.DbColl.CountDocuments(ctx, bson.M{"email": email})
	if err != nil || cnt == 0 {
		return UserNotFoundErr(err) // no user for editing
	}
	patch := bson.M{}
	if passwd != "" { // if passwd is empty we dont want to change it
		up := UserPassword(passwd)
		if !up.IsValid() {
			return InvalidUserErr(fmt.Errorf("invalid user password, Passwords are 9-12 alphanumerical characters including special symbols"))
		}
		hashStr, err := up.StringHash()
		if err != nil {
			return InvalidUserErr(err)
		}
		patch["auth"] = hashStr
	}
	if name != "" {
		if UserName(name).IsValid() {
			patch["name"] = name
		} else {
			return InvalidUserErr(fmt.Errorf("invalid user name %s", name))
		}
	}
	if telegid != int64(0) {
		patch["telegid"] = telegid
	}
	ctx, _ = context.WithCancel(context.Background())                               // if set withtimeout, 5 seconds isnt enough since generating the hash would take some time dependingon theccost
	_, err = u.DbColl.UpdateOne(ctx, bson.M{"email": email}, bson.M{"$set": patch}) // user updated

	if err != nil {
		return FailedDBQueryErr(err)
	}
	return nil
}

func (u *UsersCollection) NewUser(usr *User) error {
	// Chcking for the name
	if !UserName(usr.Name).IsValid() {
		return InvalidUserErr(fmt.Errorf("invalid name of the user"))
	}
	// Validation & hashing the password
	up := UserPassword(usr.Auth)
	if !up.IsValid() {
		return InvalidUserErr(fmt.Errorf("invalid password for user"))
	}
	hashedPasswd, err := up.StringHash()
	if err != nil {
		return InvalidUserErr(fmt.Errorf("invalid password for user"))
	}
	usr.Auth = hashedPasswd

	if !UserEmail(usr.Email).IsValid() {
		return InvalidUserErr(fmt.Errorf("invalid email for user"))
	}

	// Checking for duplicates
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	cnt, err := u.DbColl.CountDocuments(ctx, bson.M{"email": usr.Email}) // no 2 users can have the same email
	if err != nil {
		return FailedDBQueryErr(err)
	}
	if cnt != 0 {
		return DuplicateEntryErr(fmt.Errorf("User already registered"))
	}

	// Finally inserting the new user details
	_, err = u.DbColl.InsertOne(ctx, usr)
	if err != nil {
		return fmt.Errorf("failed NewUser : %s", err)
	}
	return nil
}
