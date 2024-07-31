package main

import (
	"os"
	"testing"

	"github.com/anishjain94/mongo-oplog-to-sql/constants"
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

func TestStoreCheckpoint(t *testing.T) {
	var wantFileCheckpoint int64 = 100
	var wantMongoCheckpoint uint32 = 150
	var wantMongoCheckpointDb string = "database1"

	constants.LastReadCheckpoint.SetFileCheckpoint(wantFileCheckpoint)
	constants.LastReadCheckpoint.SetMongoCheckpoint(wantMongoCheckpointDb, primitive.Timestamp{
		T: wantMongoCheckpoint,
	})

	err := saveLastRead()
	if err != nil {
		t.Fatal(err)
	}

	err = restoreCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	gotFileCheckpoint := constants.LastReadCheckpoint.GetFileCheckpoint()

	if gotFileCheckpoint != wantFileCheckpoint {
		t.Errorf("Got %v\nWant %v", gotFileCheckpoint, wantFileCheckpoint)
	}

	gotMongoCheckpoint := constants.LastReadCheckpoint.GetMongoCheckpoint(wantMongoCheckpointDb)

	if gotMongoCheckpoint.T != wantMongoCheckpoint {
		t.Errorf("Got %v\nWant %v", gotFileCheckpoint, wantFileCheckpoint)
	}
}
