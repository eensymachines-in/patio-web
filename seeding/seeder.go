package seeding

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

var (
	MongoSeeder = func(server, user, password, dbname, coll string) (Seeder, error) {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s:27017", user, password, server)))
		if err != nil {
			return nil, err
		}
		// trying the source of seeder
		return &mongoSeeder{
			client: client,
			coll:   client.Database(dbname).Collection(coll),
		}, nil
	}
)

type Seeder interface {
	DestinationEmpty() bool
	Seed(sr SeedReader) (int64, error)
	Flush() error
}

type mongoSeeder struct {
	client *mongo.Client
	coll   *mongo.Collection
}

func (ms *mongoSeeder) DestinationEmpty() bool {
	ctx := context.Background()
	cnt, _ := ms.coll.CountDocuments(ctx, bson.M{})
	return cnt == 0

}

// Flush : flushes the entire collection that is set - not to be used in production
// CAUTION: Use this function only in dev environment, use with force feed database for testing database
func (ms *mongoSeeder) Flush() error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err := ms.coll.DeleteMany(ctx, bson.M{})
	return err
}

// Seed : takes in the reader, and pushes the  data from the seed to the database
//
// sr			: reads from the json file underneath onto readResult
//
// readResult 	: typically a slice of objects onto which the reader will read onto
//
// Returns the number of records seeded & error incase if any
func (ms *mongoSeeder) Seed(sr SeedReader) (int64, error) {
	if ms.coll == nil {
		return int64(0), fmt.Errorf("source / destination of the seeder are empty/nil, cannot proceed")
	}
	// devices := []devices.Device{}
	readResult := []map[string]interface{}{}
	err := sr.Read(&readResult)
	if err != nil {
		return int64(0), err
	}
	ctx, _ := context.WithCancel(context.Background())
	for idx, item := range readResult {
		_, err := ms.coll.InsertOne(ctx, item)
		if err != nil {
			return int64(idx + 1), fmt.Errorf("error inserting seed, database could be partially seeded")
		}
	}
	cnt, err := ms.coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		return int64(0), fmt.Errorf("error getting the count of documents")
	}
	return cnt, nil
}
