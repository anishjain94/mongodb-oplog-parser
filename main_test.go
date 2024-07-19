package main

import (
	"context"
	"testing"
)

func TestMain(t *testing.T) {
	var queries []string

	inputFile := "example.json"
	outputFile := "output.sql"

	config := FlagConfig{
		InputFilePath:  &inputFile,
		OutputFilePath: &outputFile,
	}

	decodedData, err := readFile(config)
	if err != nil {
		t.Error(err)
	}

	for _, logs := range decodedData {
		queriesToAppend := GetSqlQueries(logs)
		if err != nil {
			t.Error(err)
		}
		queries = append(queries, queriesToAppend...)
	}

	displayOutput(config, queries)
}

func TestMongo(t *testing.T) {
	ctx := context.Background()
	InitializeMongoDb()
	InitializePostgres()

	var queries []string
	opLogs := GetOpLogs()

	for _, opLog := range opLogs {
		queriesToAppend := GetSqlQueries(opLog)
		queries = append(queries, queriesToAppend...)
	}

	executeQueries(&ctx, queries)

}
