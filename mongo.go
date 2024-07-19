package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func InitializeMongoDb() {
	ctx := context.Background()
	var err error

	url := fmt.Sprintf("mongodb://%s:%s@%s:%s/",
		os.Getenv("MONGODB_ROOT"),
		os.Getenv("MONGODB_PASSWORD"),
		os.Getenv("MONGODB_HOST"),
		os.Getenv("MONGODB_PORT"),
	)

	Client, err = mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		log.Fatal(err)
	}

	err = Client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func GetOpLogs() []Oplog {
	ctx := context.Background()

	coll := Client.Database(os.Getenv("MONGODB_DBNAME")).Collection("logs")
	cursor, err := coll.Find(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}

	var result []Oplog
	if err = cursor.All(ctx, &result); err != nil {
		log.Fatal(err)
	}

	return result
}
