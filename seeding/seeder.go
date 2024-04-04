package seeding

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/eensymachines-in/patio-web/devices"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

var (
	MongoSeeder = func(filepath, server, user, password, dbname, coll string) (Seeder, error) {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s:27017", user, password, server)))
		if err != nil {
			return nil, err
		}
		// trying the source of seeder
		f, err := os.Open(filepath)
		if err != nil {
			return nil, err
		}
		return &mongoSeeder{
			fp:     f,
			client: client,
			coll:   client.Database(dbname).Collection(coll),
		}, nil
	}
)

type Seeder interface {
	DestinationEmpty() bool
	Seed() (int64, error)
}

type mongoSeeder struct {
	fp     *os.File
	client *mongo.Client
	coll   *mongo.Collection
}

func (ms *mongoSeeder) DestinationEmpty() bool {
	return true
}

func (ms *mongoSeeder) Seed() (int64, error) {
	if ms.fp == nil || ms.coll == nil {
		return int64(0), fmt.Errorf("source / destination of the seeder are empty/nil, cannot proceed")
	}
	byt, err := io.ReadAll(ms.fp)
	if err != nil {
		return int64(0), err
	}
	result := []devices.Device{}
	if err := json.Unmarshal(byt, &result); err != nil {
		return int64(0), err
	}
	ctx, _ := context.WithCancel(context.Background())
	for idx, item := range result {
		_, err := ms.coll.InsertOne(ctx, item)
		if err != nil {
			// for all the documents already inserted
			return int64(idx - 1), err
		}
	}
	cnt, err := ms.coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		return int64(0), fmt.Errorf("error getting the count of documents")
	}
	return cnt, nil
}
