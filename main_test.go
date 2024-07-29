package main

import (
	"context"
	"os"
	"testing"

	"github.com/anishjain94/mongo-oplog-to-sql/constants"
	"github.com/anishjain94/mongo-oplog-to-sql/database/mongodb"
	"github.com/anishjain94/mongo-oplog-to-sql/database/postgres"
	"github.com/anishjain94/mongo-oplog-to-sql/models"
	"github.com/anishjain94/mongo-oplog-to-sql/transformer"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMain(t *testing.T) {
	var queries []string

	inputFile := "example-input.json"
	outputFile := "example-output.sql"

	config := models.FlagConfig{
		InputFilePath:  inputFile,
		OutputFilePath: outputFile,
	}
	var oplogChannel = make(chan models.Oplog)

	go func() {
		err := readFileContent(config.InputFilePath, oplogChannel)
		if err != nil {
			t.Error(err)
		}
		close(oplogChannel)
	}()

	for opLog := range oplogChannel {
		queriesToAppend := transformer.GetSqlQueries(opLog)
		queries = append(queries, queriesToAppend...)
	}
	createOutputFile(config, queries)

	_, err := os.ReadFile(config.OutputFilePath)
	if err != nil {
		t.Error(err)
	}

}

func TestMongo(t *testing.T) {
	ctx := context.Background()
	mongodb.InitializeMongoDb()
	postgres.InitializePostgres()

	var oplogChannel = make(chan models.Oplog)

	var queries []string
	go func() {
		mongodb.WatchCollection(ctx, oplogChannel)
	}()

	for opLog := range oplogChannel {
		queriesToAppend := transformer.GetSqlQueries(opLog)
		queries = append(queries, queriesToAppend...)
	}

	postgres.ExecuteQueries(ctx, queries...)

	if len(queries) == 0 {
		t.Errorf("queries not generated")
	}
}

func TestStoreCheckpoint(t *testing.T) {
	var wantFileCheckpoint int64 = 100
	var wantMongoCheckpoint uint32 = 150

	constants.LastReadCheckpoint = constants.LastReadCheckpointConfig{
		FileLastReadPosition: wantFileCheckpoint,
		MongoLastReadPosition: primitive.Timestamp{
			T: wantMongoCheckpoint,
		},
	}

	err := SaveLastRead()
	if err != nil {
		t.Fatal(err)
	}

	err = RestoreLastRead()
	if err != nil {
		t.Fatal(err)
	}

	gotFileCheckpoint := constants.LastReadCheckpoint.GetFileCheckpoint()

	if gotFileCheckpoint != wantFileCheckpoint {
		t.Errorf("Got %v\nWant %v", gotFileCheckpoint, wantFileCheckpoint)
	}

	gotMongoCheckpoint := constants.LastReadCheckpoint.GetMongoCheckpoint()

	if gotMongoCheckpoint.T != wantMongoCheckpoint {
		t.Errorf("Got %v\nWant %v", gotFileCheckpoint, wantFileCheckpoint)
	}
}
