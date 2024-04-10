package devices

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/eensymachines-in/patio-web/httperr"
	"github.com/eensymachines-in/patio/aquacfg"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

const (
	QRY_TIMEOUT = 5 * time.Second
)

type DevicesCollection struct {
	DbColl *mongo.Collection
}

func (dc *DevicesCollection) UserDevices(userID string) ([]Device, httperr.HttpErr) {
	ctx, _ := context.WithCancel(context.Background())
	result := []Device{}
	cur, err := dc.DbColl.Find(ctx, bson.M{"users": bson.M{"$elemMatch": bson.M{"$eq": userID}}})
	if err != nil {
		return result, httperr.ErrDBQuery(err)
	}
	if err := cur.All(ctx, &result); err != nil {
		return result, httperr.ErrDBQuery(err)
	}
	return result, nil
}

// TODO: meat out the implemenation here

// AddDevice : Adds a new device registration, gets the device details back along with new device id and details.

func (dc *DevicesCollection) AddDevice(d *Device) httperr.HttpErr {
	ctx, _ := context.WithCancel(context.Background())

	// Check to see if the device is already inserted
	count, err := dc.DbColl.CountDocuments(ctx, bson.M{"mac": d.Mac})
	if err != nil {
		return httperr.ErrDBQuery(err)
	}
	if count != 0 {
		return httperr.DuplicateResourceErr(fmt.Errorf("device with mac %s already registered, cannot register again", d.Mac))
	}
	if !d.Mac.IsValid() {
		logrus.WithFields(logrus.Fields{
			"mac": d.Mac,
		}).Debug("Invalid MAC id for device")
		return httperr.ErrValidation(fmt.Errorf("device mac id %s is found to be invalid", d.Mac))
	}
	ir, err := dc.DbColl.InsertOne(ctx, d)
	if err != nil {
		return httperr.ErrDBQuery(err)
	}
	oid, _ := ir.InsertedID.(primitive.ObjectID)
	sr := dc.DbColl.FindOne(ctx, bson.M{"_id": oid})
	if sr.Err() != nil {
		return httperr.ErrDBQuery(sr.Err())
	}
	if err := sr.Decode(d); err != nil {
		return httperr.ErrUnMarshal(sr.Err())
	}
	return nil
}

// EitherMacIDOrObjID: encapsulating common logic function to create and call query function
// has the logic to identify if the device is being identified by object id or the mac
// will call any device query function ahead of the identification
// basically helps to generalise /form the filtering query on the bson database
func EitherMacIDOrObjID(macOrId string, queryFn func(flt bson.M, result *Device) httperr.HttpErr) (*Device, httperr.HttpErr) {
	var flt bson.M = bson.M{}
	result := Device{} // ASK: how to get this populated incase you want to run a query to update!!
	if MacID(macOrId).IsValid() {
		flt = bson.M{"mac": macOrId}
	} else {
		oid, err := primitive.ObjectIDFromHex(macOrId)
		if err != nil { // Neither with MacID  or object id
			return nil, httperr.ErrInvalidParam(err)
		}
		flt = bson.M{"_id": oid}
	}
	return &result, queryFn(flt, &result)
}

// RemoveDevice : either send the mac or object id of the device to remove it permanently from the database
// If you have any other filter you can apply else use EitherMacIDOrObjID
func (dc *DevicesCollection) RemoveDevice(flt bson.M, result *Device) httperr.HttpErr {
	ctx, _ := context.WithTimeout(context.Background(), QRY_TIMEOUT)
	_, err := dc.DbColl.DeleteOne(ctx, flt)
	if err != nil {
		return httperr.ErrDBQuery(err)
	}
	return nil
}

// GetDevice : gets the single device given the mac or the object id hex
// If you have any other filter you can apply else use EitherMacIDOrObjID
func (dc *DevicesCollection) GetDevice(flt bson.M, result *Device) httperr.HttpErr {
	ctx, _ := context.WithTimeout(context.Background(), QRY_TIMEOUT)
	sr := dc.DbColl.FindOne(ctx, flt)
	if sr.Err() != nil {
		if errors.Is(sr.Err(), mongo.ErrNoDocuments) {
			return httperr.ErrResourceNotFound(sr.Err())
		}
		return httperr.ErrDBQuery(sr.Err())
	}
	if err := sr.Decode(result); err != nil {
		return httperr.ErrUnMarshal(sr.Err())
	}
	return nil
}

// UpdateDevice : will update the device for the users allowed
// result		: is an inout param where the users to be updated , sent in on result param
// while on its way out the result will contain the entire updated resutl of the device
// The new users slice to be updated is sent to a closure. In context of the users slice - slice of email of the user
func (dc *DevicesCollection) UpdateDevice(newUsers []string) func(flt bson.M, result *Device) httperr.HttpErr {
	return func(flt bson.M, result *Device) httperr.HttpErr {
		ctx, _ := context.WithTimeout(context.Background(), QRY_TIMEOUT)
		// AddToSet since we dont want duplicate user entries
		// $each since we DO NOT want users as array being added to array, but each element in the result.Users array to be added one at a time
		_, err := dc.DbColl.UpdateOne(ctx, flt, bson.M{"$addToSet": bson.M{"users": bson.M{"$each": newUsers}}})
		if err != nil {
			return httperr.ErrDBQuery(err)
		}
		return dc.GetDevice(flt, result) // getting the result for the updated device and repopulating the result
	}
}

// UpdateDeviceCfg : updates the configuration for the device, and sends back the updated device status, error if any
// It also sends out the configuration before updating as a back up configuration since incase downstream operations fail it shall be convenient to restore the configuration  to its original form
// Example : after updating the configuration on the server if the amqp gateway fails, since the server is single source of truth the update on the server needs to be reverted
func (dc *DevicesCollection) UpdateDeviceCfg(newCfg, backup *aquacfg.AppConfig) func(flt bson.M, result *Device) httperr.HttpErr {
	return func(flt bson.M, result *Device) httperr.HttpErr {
		ctx, _ := context.WithTimeout(context.Background(), QRY_TIMEOUT)
		// getting the back up of the device
		dc.GetDevice(flt, result)
		*backup = result.Config
		// Updating the  device for the new configuration
		_, err := dc.DbColl.UpdateOne(ctx, flt, bson.M{"$set": bson.M{"config": newCfg}})
		if err != nil {
			return httperr.ErrDBQuery(err)
		}
		// sending back the device object updated
		return dc.GetDevice(flt, result)
	}
}
