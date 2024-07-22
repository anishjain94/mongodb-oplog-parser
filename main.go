package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/anishjain94/mongo-oplog-to-sql/constants"
	"github.com/anishjain94/mongo-oplog-to-sql/database/mongodb"
	"github.com/anishjain94/mongo-oplog-to-sql/database/postgres"
	"github.com/anishjain94/mongo-oplog-to-sql/models"
	"github.com/anishjain94/mongo-oplog-to-sql/transformer"
	"github.com/joho/godotenv"
)

func init() {
	ctx := context.Background()
	err := godotenv.Load("./.env")
	if err != nil {
		log.Fatal("unable to load .env file")
	}

	mongodb.InitializeMongoDb(&ctx)
	postgres.InitializePostgres()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	flagConfig := parseFlags()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

	done := make(chan struct{}) //channel to notify when all goroutines finishes.

	go func() {
		for {
			select {
			case <-ctx.Done():
				return //exit goroutine when ctx is cancelled

			default:
				wg.Add(1)
				go func(ctx *context.Context, flagConfig *models.FlagConfig, wg *sync.WaitGroup) {
					defer wg.Done()
					if err := RunMainLogic(ctx, flagConfig); err != nil {
						log.Fatal(err.Error())
					}
				}(&ctx, flagConfig, &wg)
			}
		}
	}()

	<-sigChan
	log.Println("Shutdown signal received. Gracefully shutting down")

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("gracefully shutdown")
	case <-time.After(5 * time.Second):
		log.Println("Shutdown timed out")
	}
}

func RunMainLogic(ctx *context.Context, flagConfig *models.FlagConfig) error {
	var decodedData []models.Oplog
	var err error

	// Get input from json file or mongodb
	switch flagConfig.InputType {
	case constants.InputTypeJSON:
		decodedData, err = readFile(flagConfig.InputFilePath)
		if err != nil {
			return err
		}

	case constants.InputTypeMongoDB:
		decodedData, err = mongodb.GetOpLogs(ctx)
		if err != nil {
			return err
		}
	}

	var queries []string
	for _, logs := range decodedData {
		sqlQueries := transformer.GetSqlQueries(logs)
		if err != nil {
			log.Fatal(err)
		}
		queries = append(queries, sqlQueries...)
	}

	// Execute queries or create sql file
	switch flagConfig.OutputType {
	case constants.OutputTypeSQL:
		err = createOutputFile(*flagConfig, queries)
		if err != nil {
			return err
		}

	case constants.OutputTypeDB:
		err = postgres.ExecuteQueries(ctx, queries...)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseFlags() *models.FlagConfig {
	config := &models.FlagConfig{}

	inputType := flag.String("input-type", string(constants.InputTypeJSON), "Input type: 'json' or 'mongodb'")
	outputType := flag.String("output-type", string(constants.OutputTypeSQL), "Output type: 'sql' or 'db'")

	// Input file path (for JSON input)
	flag.StringVar(&config.InputFilePath, "i", "", "Input JSON file path (required for JSON input)")
	// Output file path (for SQL file output)
	flag.StringVar(&config.OutputFilePath, "o", "", "Output SQL file path (required for SQL file output)")

	flag.Parse()

	// Validate and set input type
	switch constants.InputType(*inputType) {
	case constants.InputTypeJSON:
		config.InputType = constants.InputTypeJSON
		if config.InputFilePath == "" {
			log.Println("Error: Input file path is required for JSON input")
			flag.Usage()
			os.Exit(1)
		}
	case constants.InputTypeMongoDB:
		config.InputType = constants.InputTypeMongoDB

	default:
		log.Printf("Error: Invalid input type '%s'\n", *inputType)
		flag.Usage()
		os.Exit(1)
	}

	// Validate and set output type
	switch constants.OutputType(*outputType) {
	case constants.OutputTypeSQL:
		config.OutputType = constants.OutputTypeSQL
		if config.OutputFilePath == "" {
			log.Println("Error: Output file path is required for SQL file output")
			flag.Usage()
			os.Exit(1)
		}
	case constants.OutputTypeDB:
		config.OutputType = constants.OutputTypeDB

	default:
		log.Printf("Error: Invalid output type '%s'\n", *outputType)
		flag.Usage()
		os.Exit(1)
	}

	return config
}

func createOutputFile(flagConfig models.FlagConfig, queries []string) error {
	file, err := os.OpenFile(flagConfig.OutputFilePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	buffer := bufio.NewWriter(file)

	// TODO: do chunk insertion into files and then flush it.
	for _, query := range queries {
		_, err := buffer.Write([]byte(query))
		if err != nil {
			return err
		}
	}

	if err := buffer.Flush(); err != nil {
		return err
	}

	return nil
}

func readFile(filePath string) ([]models.Oplog, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)

	_, err = decoder.Token()
	if err != nil {
		return nil, err
	}

	var decodedDataList []models.Oplog

	for decoder.More() {
		var decodedData models.Oplog
		err = decoder.Decode(&decodedData)
		if err != nil {
			return nil, err
		}
		decodedDataList = append(decodedDataList, decodedData)
	}
	return decodedDataList, nil
}
