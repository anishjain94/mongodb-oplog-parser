package mongodb

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/anishjain94/mongo-oplog-to-sql/constants"
	"github.com/anishjain94/mongo-oplog-to-sql/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

func InitializeMongoDb() error {
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URL")).SetDirect(true)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return err
	}

	MongoClient = client
	return nil
}

func WatchCollection(ctx context.Context, opLog chan<- models.Oplog) error {
	collection := MongoClient.Database("local").Collection("oplog.rs")
	lastReadPosition := constants.LastReadCheckpoint.GetMongoCheckpoint()
	filter := buildFilter(lastReadPosition)

	findOptions := options.Find().SetCursorType(options.TailableAwait)
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil

		default:
			if cursor.TryNext(ctx) {
				var data bson.M
				if err := cursor.Decode(&data); err != nil {
					return err
				}

				jsonData, err := json.Marshal(data)
				if err != nil {
					return err
				}

				var entry models.Oplog
				if err = json.Unmarshal(jsonData, &entry); err != nil {
					return err
				}

				select {
				case opLog <- entry:
					// Successfully sent the entry
				case <-ctx.Done():
					return fmt.Errorf("context cancelled while sending entry")
				}

				constants.LastReadCheckpoint.SetMongoCheckpoint(entry.Timestamp)
			}
		}
	}
}

func buildFilter(lastReadPosition primitive.Timestamp) bson.M {
	filter := bson.M{
		"op": bson.M{"$nin": []string{"n", "c"}},
		"$and": []bson.M{
			{"ns": bson.M{"$not": bson.M{"$regex": "^(admin|config)\\."}}},
			{"ns": bson.M{"$not": bson.M{"$eq": ""}}},
		},
	}

	if lastReadPosition.T != 0 {
		filter["$and"] = append(filter["$and"].([]bson.M), bson.M{"ts": bson.M{"$gte": lastReadPosition}})
	}

	return filter
}
