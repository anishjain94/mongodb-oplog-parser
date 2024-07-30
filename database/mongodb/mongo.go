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

func WatchCollection(ctx context.Context, opLogChannel chan<- models.Oplog, dbCollection DatabaseCollection) error {
	collection := MongoClient.Database("local").Collection("oplog.rs")
	namespace := fmt.Sprintf("%s.%s", dbCollection.DatabaseName, dbCollection.CollectionName)

	lastReadPosition := constants.LastReadCheckpoint.GetMongoCheckpoint(namespace)
	filter := buildFilter(lastReadPosition, namespace)

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
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
		case opLogChannel <- entry:
			constants.LastReadCheckpoint.SetMongoCheckpoint(namespace, entry.Timestamp)

		case <-ctx.Done():
			close(opLogChannel)
			return fmt.Errorf("context cancelled while sending entry")
		}
	}

	if err := cursor.Err(); err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil
}

func buildFilter(lastReadPosition primitive.Timestamp, namespace string) bson.M {
	filter := bson.M{
		"op": bson.M{"$nin": []string{"n", "c"}},
		"$and": []bson.M{
			{"ns": bson.M{"$not": bson.M{"$regex": "^(admin|config)\\."}}},
			{"ns": bson.M{"$not": bson.M{"$eq": ""}}},
			{"ns": namespace},
		},
	}

	if lastReadPosition.T != 0 {
		filter["$and"] = append(filter["$and"].([]bson.M), bson.M{"ts": bson.M{"$gte": lastReadPosition}})
	}

	return filter
}

func ListAllCollectionsAndDatabase(ctx context.Context) ([]DatabaseCollection, error) {
	var dbCollections []DatabaseCollection

	filter := bson.M{
		"name": bson.M{
			"$nin": []string{"admin", "config", "local"},
		},
	}

	databaseNames, err := MongoClient.ListDatabaseNames(ctx, filter, nil)
	if err != nil {
		return nil, err
	}

	for _, databaseName := range databaseNames {
		collections, err := MongoClient.Database(databaseName).ListCollectionNames(ctx, filter, nil)
		if err != nil {
			return nil, err
		}

		for _, collection := range collections {
			dbCollections = append(dbCollections, DatabaseCollection{
				DatabaseName:   databaseName,
				CollectionName: collection,
			})
		}
	}

	return dbCollections, nil
}
