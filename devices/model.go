package devices

import (
	"regexp"

	"github.com/eensymachines-in/patio/aquacfg"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MacID string // string alias for device mac ids - validation

func (m MacID) IsValid() bool {
	// https://stackoverflow.com/questions/4260467/what-is-a-regular-expression-for-a-mac-address
	// delmiting character can be - or :
	r := regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)
	return r.MatchString(string(m))
}

type Device struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Mac      MacID              `bson:"mac" json:"mac"`           // unique mac id of the device
	Name     string             `bson:"name" json:"name"`         // device name, preferrably unique
	Location string             `bson:"location" json:"location"` // google location co-oridinates
	Make     string             `bson:"make" json:"make"`         // platform or make of device
	Users    []string           `bson:"users" json:"users"`       // list of authorized users that have direct access to the devices
	Config   aquacfg.AppConfig  `bson:"config" json:"config"`     // application configuration as stored in the database
}
