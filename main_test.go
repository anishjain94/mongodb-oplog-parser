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

	inputFile := "example.json"
	outputFile := "output.sql"

	config := models.FlagConfig{
		InputFilePath:  inputFile,
		OutputFilePath: outputFile,
	}

	decodedData, err := readFile(config.InputFilePath)
	if err != nil {
		t.Error(err)
	}

	for _, logs := range decodedData {
		queriesToAppend := transformer.GetSqlQueries(logs)
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
	opLogs, _ := mongodb.GetOpLogs(&ctx)

	for _, opLog := range opLogs {
		queriesToAppend := transformer.GetSqlQueries(opLog)
		queries = append(queries, queriesToAppend...)
	}

	postgres.ExecuteQueries(&ctx, queries...)
}

func TestRunMainLogic(t *testing.T) {
	ctx := context.Background()
	mongodb.InitializeMongoDb(&ctx)

	RunMainLogic(&ctx, &models.FlagConfig{
		InputType:      constants.InputTypeMongoDB,
		OutputType:     constants.OutputTypeSQL,
		OutputFilePath: "temp.sql",
	})
}
