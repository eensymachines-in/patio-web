package auth

import (
	"regexp"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type UserRole uint8

const (
	SuperUser UserRole = iota
	Admin
	EndUser
	Guest
)

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
	passRegex := regexp.MustCompile(`^[a-zA-Z\s]+$`) // password is 9-12 characters
	return passRegex.MatchString(string(un))
}

type UserEmail string

func (ue UserEmail) IsValid() bool {
	passRegex := regexp.MustCompile(`^[a-zA-Z0-9]+[_.-]*[a-zA-Z0-9]*@[a-zA-Z0-9]+[.]{1}[a-zA-Z0-9]{2,}$`) // password is 9-12 characters
	return passRegex.MatchString(string(ue))
}

// User : any user in the system, can be authenticated against database
type User struct {
	Id      primitive.ObjectID `bson:"_id,omitempty" json:"id"` // omit empty to indicate empty when marshalling and inserting
	Name    string             `bson:"name" json:"name"`
	Email   string             `bson:"email" json:"email"`
	Role    UserRole           `bson:"role" json:"role"`
	TelegID int64              `bson:"telegid" json:"telegid"`
	Auth    string             `bson:"auth"`
}
