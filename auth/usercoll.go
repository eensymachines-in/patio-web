package auth

/* =========================
project 		: ipatio-web
date			: MArch` 2024
author			: kneerunjun@gmail.com
Copyrights		: Eensy Machines
About			: Actual CRUD operation engine, does database operations, & fitting the appropriate httperr to send back on the way out over http. Does not connect to the database but uses the collection of an already connected database to fire queries.
============================*/
import (
	"context"
	"fmt"
	"time"

	"github.com/eensymachines-in/patio-web/httperr"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
)

type UsersCollection struct {
	DbColl *mongo.Collection
}

// Authenticate : will compare the email id against the hash of the password, upon success will sedn back the auth token.
// Such a token is replaced on usr.Auth on its way back a result
// This only if the user exists, else Error is returned.
//
//
/*
	usr := User{Email:"johndoe@gmail.com", Auth: "ClearTextPassword"}
	mongoClient := val.(*mongo.Client)
	uc := auth.UsersCollection{DbColl: mongoClient.Database("dbname").Collection("collname")}
	err :=uc.Authenticate(&usr)
	if err !=nil{
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
				"stack": "location-of-calling stack",
			}))
			return
		}
	}
	c.AbortWithStatusJSON(http.StatusOK, usr) // no error - user authenticated

*/
func (u *UsersCollection) Authenticate(usr *User) httperr.HttpErr {
	// usr := User{}
	ctx, _ := context.WithCancel(context.Background())

	count, err := u.DbColl.CountDocuments(ctx, bson.M{"email": usr.Email})
	if dbe := FailedDBQueryErr(err); dbe != nil {
		return dbe
	}
	if count != 1 {
		return UserNotFoundErr(fmt.Errorf("failed to get user with email %s", usr.Email))
	}
	clearTextPass := usr.Auth // before unmarshalling the user from the database, getting the cleartext password
	if err := FailedDBQueryErr(u.DbColl.FindOne(ctx, bson.M{"email": usr.Email}).Decode(usr)); err != nil {
		return err
	}
	hash := []byte(usr.Auth)
	if err := MismatchPasswdErr(bcrypt.CompareHashAndPassword(hash, []byte(clearTextPass))); err != nil {
		return err
	}

	// generate new jwt for this login
	tok := jwt.New(jwt.SigningMethodHS256) // this signing method demands key of certain type
	claims := tok.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(10 * time.Minute)
	claims["authorized"] = true
	claims["user"] = usr.Email
	claims["user_role"] = usr.Role
	usr.AuthTok, err = tok.SignedString([]byte("33n5ymach1ne5")) // []byte is ok since signing method is SigningMethodHS256
	if e := AuthTokenErr(err); e != nil {
		return e
	}
	return nil
}

// EditUser: Can edit a few fields of the user in the database, except the email.
// Typically used in PUT requests
// If the field to change is not set, will NOT update - omit empty
//
/*
	usr := User{Email:"johndoe@gmail.com", Auth: "ClearTextPassword"}
	mongoClient := val.(*mongo.Client)
	uc := auth.UsersCollection{DbColl: mongoClient.Database("dbname").Collection("collname")}
	err :=uc.EditUser(usr.Email, usr.Name, usr.Auth, usr.TelegID)
	if err !=nil{
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
				"stack": "location-of-calling stack",
			}))
			return
		}
	}
	c.AbortWithStatusJSON(http.StatusOK, usr) // no error - user authenticated
*/
func (u *UsersCollection) EditUser(email string, name, passwd string, telegid int64) httperr.HttpErr {
	ctx, _ := context.WithCancel(context.Background())
	// Figuring out if the identifying param is email / id hex
	var flt bson.M
	if UserEmail(email).IsValid() {
		flt = bson.M{"email": email}
	} else {
		hexID, err := primitive.ObjectIDFromHex(email)
		if err != nil {
			return InvalidUserErr(fmt.Errorf("User identifier is invalid, check and send again"))
		}
		flt = bson.M{"_id": hexID}
	}
	cnt, err := u.DbColl.CountDocuments(ctx, flt)
	if err != nil {
		return FailedDBQueryErr(err) // no user for editing
	}
	if cnt == 0 {
		return UserNotFoundErr(fmt.Errorf("failed to get user %s", email))
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
	ctx, _ = context.WithCancel(context.Background())            // if set withtimeout, 5 seconds isnt enough since generating the hash would take some time dependingon theccost
	_, err = u.DbColl.UpdateOne(ctx, flt, bson.M{"$set": patch}) // user updated

	if err != nil {
		return FailedDBQueryErr(err)
	}
	return nil
}

// NewUser : Can insert new user account if it isnt already inserted.
// Email of the account serves as the unique identifier for the account. No 2 accounts with the same email can exists in the same database.
// Email, password, and Name all have regex validation checks - anyone fails it will not insert the account and return 400.
// Error ireturned is directly compatible with httperr.HttpErrOrOkDispatch
//
/*
	usr := User{Email:"johndoe@gmail.com", Auth: "ClearTextPassword", Name: "John Doe", TelegID: 6645654654}
	mongoClient := val.(*mongo.Client)
	uc := auth.UsersCollection{DbColl: mongoClient.Database("dbname").Collection("collname")}
	err :=uc.NewUser(usr) // of the type httperr.HttpErr
	if err !=nil{
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
				"stack": "location-of-calling stack",
			}))
			return
		}
	}
	c.AbortWithStatusJSON(http.StatusOK, usr) // no error - user authenticated
*/
func (u *UsersCollection) NewUser(usr *User) httperr.HttpErr {
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
	ctx, _ := context.WithCancel(context.Background())
	cnt, err := u.DbColl.CountDocuments(ctx, bson.M{"email": usr.Email}) // no 2 users can have the same email
	if err != nil {
		return FailedDBQueryErr(err)
	}
	if cnt != 0 {
		return DuplicateEntryErr(fmt.Errorf("User already registered"))
	}

	// Finally inserting the new user details
	insertResult, err := u.DbColl.InsertOne(ctx, usr)
	usr.Id = insertResult.InsertedID.(primitive.ObjectID) // newly inserted document id
	if err != nil {
		return FailedDBQueryErr(fmt.Errorf("failed NewUser : %s", err))
	}
	return nil
}

// DeleteUser : given the email/id this can delete the account. Once deleted the account cannot be recovered.
// Incase the account isnt found throws NotFoundErr
// It can figure out if the email or ID is used for addressing the account to be deleted
//
/*
	mongoClient := val.(*mongo.Client)
	uc := auth.UsersCollection{DbColl: mongoClient.Database("dbname").Collection("collname")}
	err :=uc.DeleteUser("johndoe@gmail.com") // of the type httperr.HttpErr
	if err !=nil{
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
				"stack": "location-of-calling stack",
			}))
			return
		}
	}
	c.AbortWithStatusJSON(http.StatusOK, usr) // no error - user authenticated
*/
func (u *UsersCollection) DeleteUser(emailOrID string) httperr.HttpErr {
	ctx, _ := context.WithCancel(context.Background())
	var flt bson.M
	if !UserEmail(emailOrID).IsValid() { // if its email or hex object id
		// return InvalidUserErr(fmt.Errorf("invalid email for user"))
		oid, err := primitive.ObjectIDFromHex(emailOrID)
		if err != nil {
			return InvalidUserErr(err) // id of the user is invalid
		}
		flt = bson.M{"_id": oid}
	} else {
		flt = bson.M{"email": emailOrID}
	}
	delResult, err := u.DbColl.DeleteOne(ctx, flt)
	if err != nil {
		return FailedDBQueryErr(fmt.Errorf("failed DeleteUser : %s", err))
	}
	if delResult.DeletedCount == 0 {
		return UserNotFoundErr(fmt.Errorf("user account %s was not found", emailOrID))
	}
	return nil
}
