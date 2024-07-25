package main

import (
	"context"
	"reflect"
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
	var oplogChannel = make(chan models.Oplog)

	err := readFileContent(config.InputFilePath, oplogChannel)
	if err != nil {
		t.Error(err)
	}

	for opLog := range oplogChannel {
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
	mongodb.InitializeMongoDb()
	postgres.InitializePostgres()

	var oplogChannel = make(chan models.Oplog)

	var queries []string
	mongodb.WatchCollection(ctx, oplogChannel)

	for opLog := range oplogChannel {
		queriesToAppend := transformer.GetSqlQueries(opLog)
		queries = append(queries, queriesToAppend...)
	}

	postgres.ExecuteQueries(ctx, queries...)
}

func TestRunMainLogic(t *testing.T) {
	ctx := context.Background()
	mongodb.InitializeMongoDb()

	parseOplog(ctx, &models.FlagConfig{
		InputType:     constants.InputTypeJSON,
		InputFilePath: "example-input.json",

		OutputType:     constants.OutputTypeSQL,
		OutputFilePath: "temp.sql",
	})
}

func TestStoreCheckpoint(t *testing.T) {
	columns := []string{"col1", "col2", "col3"}

	globalConfig := constants.GetGlobalVariables()
	globalConfig.CreateSchemaQuery.Set("createSchema", true)
	globalConfig.CreateTableQuery.Set("createTable", true)
	globalConfig.TableColumnName.Set("tablecolumns", columns)

	err := constants.StoreCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	err = constants.RestoreCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	restoredConfig := constants.GetGlobalVariables()

	if val, _ := restoredConfig.CreateSchemaQuery.Get("createSchema"); !val {
		t.Errorf("createSchema in restoredconfig not found")
	}

	if val, _ := restoredConfig.CreateTableQuery.Get("createTable"); !val {
		t.Errorf("createSchema in restoredconfig not found")
	}

	if val, exists := restoredConfig.TableColumnName.Get("tablecolumns"); !exists && !reflect.DeepEqual(columns, val) {
		t.Errorf("createSchema in restoredconfig not found")
	}
}
