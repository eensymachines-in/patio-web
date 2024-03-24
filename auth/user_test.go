package auth

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
)

func TestEditUser(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://eensyaquap-dev:33n5y+4dm1n@65.20.79.167:57017"))
	assert.Nil(t, err, "Unexpected error connecting to database")
	assert.NotNil(t, client, "Unexpected nil connection after connecting to database")
	uc := UsersCollection{DbColl: client.Database("aquaponics").Collection("users")}

	// Test setup
	data := &User{Name: "Niranjan", Email: "niranjan_Awati@gmail.com", Role: SuperUser, TelegID: int64(53454353), Auth: "9characters"}
	uc.NewUser(data) // setting up the test

	edit := User{Email: data.Email, Name: "kneerunjun", Auth: "oZ7@mNav6U", TelegID: int64(6564565464)}
	err = uc.EditUser(data.Email, edit.Name, edit.Auth, edit.TelegID)
	assert.Nil(t, err, "Unexpected error when editing user")

	result := User{}
	uc.DbColl.FindOne(ctx, bson.M{"email": edit.Email}).Decode(&result)
	assert.Equal(t, edit.Name, result.Name, "Name found mismatchng")
	assert.Equal(t, edit.TelegID, result.TelegID, "Name found mismatchng")
	compare := bcrypt.CompareHashAndPassword([]byte(result.Auth), []byte(edit.Auth))
	assert.Nil(t, compare, "Unexpected  error when comparing the passwortds")

	// trying to modify only the name
	edit = User{Email: data.Email, Name: "niryabhadkao"}
	err = uc.EditUser(data.Email, edit.Name, "", int64(0))
	assert.Nil(t, err, "Unexpected error when updating the user name")
	uc.DbColl.FindOne(ctx, bson.M{"email": edit.Email}).Decode(&result)
	assert.Equal(t, edit.Name, result.Name, "Name hasnt been updated, unexpected")

	// trying to alter only the password
	edit = User{Email: data.Email, Auth: "rC8%$bzQmnp*"}
	err = uc.EditUser(data.Email, "", edit.Auth, int64(0))
	assert.Nil(t, err, "Unexpected error when updating the user password ")
	uc.DbColl.FindOne(ctx, bson.M{"email": edit.Email}).Decode(&result)
	compare = bcrypt.CompareHashAndPassword([]byte(result.Auth), []byte(edit.Auth))
	assert.Nil(t, compare, "Password hasnt been updated")

	// Altering only the TelegramId
	edit = User{Email: data.Email, Auth: "", TelegID: int64(56457587)}
	err = uc.EditUser(data.Email, "", "", edit.TelegID)
	assert.Nil(t, err, "Unexpected error when updating the user telegram id")
	uc.DbColl.FindOne(ctx, bson.M{"email": edit.Email}).Decode(&result)
	assert.Equal(t, edit.TelegID, result.TelegID, "Telegram id hasnt been updated, unexpected")

	// TEST: user not found ...
	edit = User{Email: "anonymous@hagga.com", Auth: ""}
	err = uc.EditUser(edit.Email, "", "", edit.TelegID)
	assert.NotNil(t, err, "Uexpected nil error when user not found")
	// TEST: invalid password ..
	edit = User{Email: data.Email, Auth: ">>????????"}
	err = uc.EditUser(edit.Email, "", edit.Auth, edit.TelegID)
	assert.NotNil(t, err, "Uexpected nil error when user password is invalid")
	// TEST: invalid name
	edit = User{Email: data.Email, Name: "4534555345", Auth: "oZ7@mNav6U"} // invalid name
	// Invalid  name is not empty name, empty name denotes no changes in name edit
	err = uc.EditUser(edit.Email, edit.Name, edit.Auth, edit.TelegID)
	assert.NotNil(t, err, "Uexpected nil error when user name is invalid")

	// Cleaning up the test
	uc.DbColl.DeleteMany(ctx, bson.M{})
}

func TestNewUser(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://eensyaquap-dev:33n5y+4dm1n@65.20.79.167:57017"))
	assert.Nil(t, err, "Unexpected error connecting to database")
	assert.NotNil(t, client, "Unexpected nil connection after connecting to database")
	uc := UsersCollection{DbColl: client.Database("aquaponics").Collection("users")}
	data := []User{
		{Name: "Niranjan", Email: "niranjan_Awati@gmail.com", Role: SuperUser, TelegID: int64(53454353), Auth: "9characters"},
		{Name: "Awati", Email: "kneerunjun@gmail.com", Role: Admin, TelegID: int64(53454353), Auth: "someRandom@2"},
		{Name: "NiryaBhadkao", Email: "bhadkao.nirya@gmail.com", Role: EndUser, TelegID: int64(126552564), Auth: "ter$$dfgdg"},
		{Name: "Niranjan", Email: "aaichagho.tujhya@gmail.com", Role: Guest, TelegID: int64(53454353), Auth: "hoehoe@56778"},
	}
	for _, u := range data {
		err := uc.NewUser(&u)
		assert.Nil(t, err, "Unexpected error when inserting new user")
	}
	c, _ := uc.DbColl.CountDocuments(ctx, bson.M{})
	assert.Equal(t, int64(4), c, "Expected 4 document in the collection")

	// Now for the negative tests

	notOkData := []User{
		{Name: "Niranjan", Email: "niranjan_Awati@gmail.com", Role: SuperUser, TelegID: int64(53454353), Auth: "9characters"}, // duplicate entry
		{Name: "", Email: "niranjan.awati@hotmail.com", Role: SuperUser, TelegID: int64(53454353), Auth: "9characters"},       // empty name
		{Name: "Niranjan", Email: "john.doe@gmail.com", Role: SuperUser, TelegID: int64(53454353), Auth: ""},                  // empty passwor
		{Name: "Niranjan", Email: ".doe@gmail.com", Role: SuperUser, TelegID: int64(53454353), Auth: "9characters"},           // invalid email

	}
	for _, u := range notOkData {
		err := uc.NewUser(&u)
		assert.NotNil(t, err, "Unexpected nil error, was expecting an error")
	}
	uc.DbColl.DeleteMany(ctx, bson.M{}) // cleaning up the test..
}

func TestPassRegex(t *testing.T) {
	okdata := []string{
		"jun%41993",
		"runjuN%41993",
		"awesome@2803",
		"gruesome@41",
		"0123456789",
		"_!@#$%^&*-",
	}
	passRegex := regexp.MustCompile(`^[0-9a-zA-Z_!@#$%^&*-]{9,12}$`)
	for _, d := range okdata {
		result := passRegex.MatchString(d)
		assert.True(t, result, "Regex test failed unexpectedly")
	}
	notokdata := []string{
		"",                // length not enough
		"(..)(..)",        // disallowed characters
		"??>>",            // length and disallowed characters
		">>>>>>>>>>>>>>>", // length and disallowed characters
		"ttrt",            // length
	}
	for _, d := range notokdata {
		result := passRegex.MatchString(d)
		assert.False(t, result, "Regex test failed unexpectedly")
	}
}

func TestEmailRegex(t *testing.T) {
	okdata := []string{
		"niranjan_awati@gmail.com",
		"kneerunjun@gmail.com",
		"niranjan.awati@gmail.com",
		"niranjan-awati@gmail.com",
		"n-awati@gmail.com",
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9]+[_.-]*[a-zA-Z0-9]*@[a-zA-Z0-9]+[.]{1}[a-zA-Z0-9]{2,}$`)
	for _, d := range okdata {
		result := emailRegex.MatchString(d)
		assert.True(t, result, "Regex test failed unexpectedly")
	}
	notokdata := []string{
		".awti@gmail.com",        // missing first name
		"",                       // empty email
		"niranjanawati",          // not an email at all
		"nirnajan_awati@gmail",   // domain name not complete
		"nirnajan_awati@gmail.",  // domain name not complete
		"nirnajan_awati@gmail.1", // domain name not complete
	}
	for _, d := range notokdata {
		result := emailRegex.MatchString(d)
		assert.False(t, result, "Regex test failed unexpectedly")
	}
}

func TestUserInsert(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://eensyaquap-dev:33n5y+4dm1n@65.20.79.167:57017"))
	assert.Nil(t, err, "Unexpected error connecting to database")
	assert.NotNil(t, client, "Unexpected nil connection after connecting to database")

	uc := UsersCollection{DbColl: client.Database("aquaponics").Collection("users")}
	err = uc.NewUser(&User{Name: "Niranjan Awati", Email: "kneerunjun@gmailcom", Role: SuperUser, TelegID: int64(0), Auth: "280382"})
	assert.Nil(t, err, "Unexpected error when inserting new user")
}
