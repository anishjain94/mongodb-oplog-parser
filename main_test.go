package main

import (
	"context"
	"testing"

	"github.com/anishjain94/mongo-oplog-to-sql/constants"
	"github.com/anishjain94/mongo-oplog-to-sql/database/mongodb"
	"github.com/anishjain94/mongo-oplog-to-sql/database/postgres"
	"github.com/anishjain94/mongo-oplog-to-sql/models"
	"github.com/anishjain94/mongo-oplog-to-sql/transformer"
)

func TestMain(t *testing.T) {
	var queries []string

	inputFile := "example-input.json"
	outputFile := "example-output.sql"

	config := models.FlagConfig{
		InputFilePath:  inputFile,
		OutputFilePath: outputFile,
	}

	err := readFileContent(config.InputFilePath, models.OpLogChannel)
	if err != nil {
		t.Error(err)
	}

	for opLog := range models.OpLogChannel {
		queriesToAppend := transformer.GetSqlQueries(opLog)
		if err != nil {
			t.Error(err)
		}
		queries = append(queries, queriesToAppend...)
	}

	createOutputFile(config, queries)
}

func TestMongo(t *testing.T) {
	ctx := context.Background()
	mongodb.InitializeMongoDb(&ctx)
	postgres.InitializePostgres()

	var queries []string
	mongodb.WatchCollection(&ctx, models.OpLogChannel)

	for opLog := range models.OpLogChannel {
		queriesToAppend := transformer.GetSqlQueries(opLog)
		queries = append(queries, queriesToAppend...)
	}

	postgres.ExecuteQueries(&ctx, queries...)
}

func TestRunMainLogic(t *testing.T) {
	ctx := context.Background()
	mongodb.InitializeMongoDb(&ctx)

	RunMainLogic(&ctx, &models.FlagConfig{
		InputType:     constants.InputTypeJSON,
		InputFilePath: "example-input.json",

		OutputType:     constants.OutputTypeSQL,
		OutputFilePath: "temp.sql",
	})
}
