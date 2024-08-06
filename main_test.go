package main

import (
	"context"
	"sync"
	"testing"

	"github.com/anishjain94/mongo-oplog-to-sql/constants"
	"github.com/anishjain94/mongo-oplog-to-sql/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// func TestMain(t *testing.T) {
// 	err := fileOplogReading()
// 	if err != nil {
// 		t.Error(err)
// 	}
// }

func fileOplogReading() error {
	ctx := context.Background()

	inputFile := "example-input.json"
	outputFile := "example-output.sql"

	flagConfig := &models.FlagConfig{
		InputFilePath:  inputFile,
		InputType:      constants.InputTypeJSON,
		OutputFilePath: outputFile,
		OutputType:     constants.OutputTypeSQL,
	}
	var wg sync.WaitGroup

	oplogChannels := readFromSource(ctx, flagConfig)

	wg.Add(len(oplogChannels))
	for _, ch := range oplogChannels {
		go func(channel chan models.OplogEntry) {
			defer wg.Done()
			processOplog(ctx, &models.ProcessOplog{
				Channel:    channel,
				FlagConfig: flagConfig,
			})
		}(ch)
	}

	wg.Wait()
	return nil
}

func BenchmarkFileOplog(b *testing.B) {
	for i := 0; i < b.N; i++ {
		err := fileOplogReading()
		if err != nil {
			b.Fatal(err)
		}
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
