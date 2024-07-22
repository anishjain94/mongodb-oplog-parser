package mongodb

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/anishjain94/mongo-oplog-to-sql/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

func InitializeMongoDb(ctx *context.Context) {
	var err error

	url := fmt.Sprintf("mongodb://%s:%s@%s:%s/",
		os.Getenv("MONGODB_ROOT"),
		os.Getenv("MONGODB_PASSWORD"),
		os.Getenv("MONGODB_HOST"),
		os.Getenv("MONGODB_PORT"),
	)

	MongoClient, err = mongo.Connect(*ctx, options.Client().ApplyURI(url))
	if err != nil {
		log.Fatal(err)
	}

	err = MongoClient.Ping(*ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func GetOpLogs(ctx *context.Context) ([]models.Oplog, error) {
	coll := MongoClient.Database(os.Getenv("MONGODB_DBNAME")).Collection("oplog.rs") //change it to - local.oplog.rs
	cursor, err := coll.Find(*ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	var result []models.Oplog
	if err = cursor.All(*ctx, &result); err != nil {
		return nil, err
	}

	return result, nil
}
