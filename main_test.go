package main

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/anishjain94/mongo-oplog-to-sql/constants"
	"github.com/anishjain94/mongo-oplog-to-sql/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMain(t *testing.T) {
	err := fileOplogReading()
	if err != nil {
		t.Error(err)
	}
}

func fileOplogReading() error {
	ctx := context.Background()

	inputFile := "example-input-oplog.json"
	outputFile := "example-output.sql"

	config := models.FlagConfig{
		InputFilePath:  inputFile,
		OutputFilePath: outputFile,
	}

	oplogChannel, err := readFileContent(ctx, config.InputFilePath)
	if err != nil {
		log.Fatal(err)
	}

	for i := range oplogChannel {
		go processOplog(ctx, &models.ProcessOplog{
			Channel:    oplogChannel[i],
			FlagConfig: &config,
		})
	}

	_, err = os.ReadFile(config.OutputFilePath)
	if err != nil {
		return err
	}
	return nil
}

func BenchmarkFileOplog(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fileOplogReading()
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
