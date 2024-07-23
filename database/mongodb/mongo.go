package mongodb

import (
	"context"
	"encoding/json"
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

	MongoClient, err = mongo.Connect(*ctx, options.Client().SetDirect(true).ApplyURI(url))
	if err != nil {
		log.Fatal(err)
	}

	defer MongoClient.Disconnect(*ctx)
	err = MongoClient.Ping(*ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
}

// TODO: handle event triggers to make it continously read
func GetOpLogs(ctx *context.Context) ([]models.Oplog, error) {
	collection := MongoClient.Database(os.Getenv("MONGODB_DBNAME")).Collection("oplog.rs")
	cursor, err := collection.Find(*ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	var result []models.Oplog
	if err = cursor.All(*ctx, &result); err != nil {
		return nil, err
	}

	for {
		select {
		case <-(*ctx).Done():
			return nil, nil // cancelled called.
		default:
			if cursor.TryNext(context.TODO()) {
				var data bson.M
				if err := cursor.Decode(&data); err != nil {
					panic(err)
				}

				jsonData, err := json.Marshal(data)
				if err != nil {
					panic(err)
				}

				var entry models.Oplog
				err = json.Unmarshal(jsonData, &entry)
				if err != nil {
					panic(err)
				}
				result = append(result, entry)
			}

			if err := cursor.Err(); err != nil {
				return nil, err
			}
			return result, nil
		}
	}
}

func WatchCollection(ctx *context.Context, opLog chan<- models.Oplog) error {
	collection := MongoClient.Database("local").Collection("oplog.rs")

	stream, err := collection.Watch(*ctx, mongo.Pipeline{})
	if err != nil {
		return err
	}

	defer stream.Close(*ctx)

	for stream.Next(*ctx) {

		var data bson.M
		if err := stream.Decode(&data); err != nil {
			panic(err)
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}

		var entry models.Oplog
		err = json.Unmarshal(jsonData, &entry)
		if err != nil {
			panic(err)
		}

		opLog <- entry

	}
	return nil
}
