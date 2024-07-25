package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
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
	err := godotenv.Load("./.env")
	if err != nil {
		log.Fatal("unable to load .env file")
	}

	mongodb.InitializeMongoDb()
	postgres.InitializePostgres()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ticker := time.NewTicker(5 * time.Second)

	var wg sync.WaitGroup
	flagConfig := parseFlags()
	done := make(chan struct{}) //channel to notify when all goroutines finishes.

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("ctx cancelled from main func")
				return //exit goroutine when ctx is cancelled

			case <-ticker.C:
				// TODO: overlap with multiple goroutines
				wg.Add(1)
				go func() {
					defer wg.Done()
					RunMainLogic(&ctx, flagConfig) //TODO: rename
				}()
			}
		}
	}()

	<-sigChan
	log.Println("Shutdown signal received. Gracefully shutting down")
	cancel()

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

// TODO: Define variables where they are used and not at other places.
// TODO: check if pointer ctx is used.

// TODO: 2 approacehs
// approach 1 -> close the reader when they read and hit end of file / end of stream.
// approach 2 -> always keep the producder and the reader alive.

func RunMainLogic(ctx *context.Context, flagConfig *models.FlagConfig) {
	// Get input from jsonfile/mongodb and stream that to a channel
	go func() {
		switch flagConfig.InputType {
		case constants.InputTypeJSON:
			err := readFileContent(flagConfig.InputFilePath, models.OpLogChannel)
			if err != nil {
				log.Fatal(err)
			}

		case constants.InputTypeMongoDB:
			mongodb.WatchCollection(ctx, models.OpLogChannel)
		}
	}()

	// for {
	select {
	case <-(*ctx).Done():
		fmt.Println("cancel propogated to run main logic.")
		break

	case opLog := <-models.OpLogChannel:
		sqlQueries := transformer.GetSqlQueries(opLog)

		switch flagConfig.OutputType {
		case constants.OutputTypeSQL:
			err := createOutputFile(*flagConfig, sqlQueries)
			if err != nil {
				log.Fatalln(err)
			}

		case constants.OutputTypeDB:
			err := postgres.ExecuteQueries(ctx, sqlQueries...)
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
	// }
}

func parseFlags() *models.FlagConfig {
	config := &models.FlagConfig{}

	inputType := flag.String("input-type", string(constants.InputTypeJSON), "Input type: 'json' or 'mongodb'")
	outputType := flag.String("output-type", string(constants.OutputTypeSQL), "Output type: 'sql' or 'db'")

	// Input file path (for JSON input)
	flag.StringVar(&config.InputFilePath, "i", "example-input.json", "Input JSON file path (required for JSON input)")
	// Output file path (for SQL file output)
	flag.StringVar(&config.OutputFilePath, "o", "example-output.sql", "Output SQL file path (required for SQL file output)")

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
	file, err := os.OpenFile(flagConfig.OutputFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
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

func readFileContent(filePath string, channel chan<- models.Oplog) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(constants.FileLastReadPosition, io.SeekStart)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)
	if constants.FileLastReadPosition == 0 {
		_, err = decoder.Token()
		if err != nil {
			return err
		}
	}

	for decoder.More() {
		var decodedData models.Oplog
		err = decoder.Decode(&decodedData)
		if err != nil {
			return err
		}
		constants.FileLastReadPosition, err = file.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}

		channel <- decodedData
	}

	return nil
}
