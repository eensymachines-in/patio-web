package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
)

type UsersCollection struct {
	DbColl *mongo.Collection
}

// Authenticate : will compare the email id against the password and if succecss sends back the authentication token
func (u *UsersCollection) Authenticate(email, pass string) (tokenStr string, e error) {
	usr := User{}
	ctx, _ := context.WithCancel(context.Background())

	count, err := u.DbColl.CountDocuments(ctx, bson.M{"email": email})
	if err != nil {
		e = FailedDBQueryErr(err)
		return // getting the user from the database
	}
	if count != 1 {
		e = UserNotFoundErr(fmt.Errorf("failed to get user with email %s", email))
		return
	}

	err = u.DbColl.FindOne(ctx, bson.M{"email": email}).Decode(&usr)
	if err != nil {
		e = FailedDBQueryErr(err)
		return // getting the user from the database
	}
	hash := []byte(usr.Auth)
	err = bcrypt.CompareHashAndPassword(hash, []byte(pass))
	if err != nil {
		e = MismatchPasswdErr(err)
		return // passwordd did not match
	}

	// generate new jwt for this login
	tok := jwt.New(jwt.SigningMethodHS256) // this signing method demands key of certain type
	claims := tok.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(10 * time.Minute)
	claims["authorized"] = true
	claims["user"] = email
	claims["user_role"] = usr.Role
	tokenStr, err = tok.SignedString([]byte("33n5ymach1ne5")) // []byte is ok since signing method is SigningMethodHS256
	if err != nil {
		e = AuthTokenErr(err) // error generating token
		tokenStr = ""
		return
	}
	return
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
	insertResult, err := u.DbColl.InsertOne(ctx, usr)
	usr.Id = insertResult.InsertedID.(primitive.ObjectID).Hex() // newly inserted document id
	if err != nil {
		return fmt.Errorf("failed NewUser : %s", err)
	}
	return nil
}
