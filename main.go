package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"hash/fnv"
	"io"
	"log"
	"os"
	"os/signal"
	"slices"
	"strings"
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

	err = restoreCheckpoint()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	flagConfig := parseFlags()
	done := make(chan struct{})

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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutdown signal received. Gracefully shutting down")
	cancel()

	go func() {
		wg.Wait()
		close(done)
	}()

	if err := saveLastRead(); err != nil {
		log.Panic(err)
	}

	select {
	case <-done:
		log.Println("gracefully shutdown")
	case <-time.After(5 * time.Second):
		log.Println("Shutdown timed out")
	}
}

func processOplog(ctx context.Context, processOplog *models.ProcessOplog) {
	for {
		select {
		case <-ctx.Done():
			return

		case oplog := <-processOplog.Channel:
			sqlQueries := transformer.GetSqlQueries(oplog)

			switch processOplog.FlagConfig.OutputType {
			case constants.OutputTypeSQL:
				if err := createOutputFile(*processOplog.FlagConfig, sqlQueries); err != nil {
					log.Fatalln(err)
				}

			case constants.OutputTypeDB:
				postgres.InitializePostgres()
				if err := postgres.ExecuteQueries(ctx, sqlQueries...); err != nil {
					log.Fatalln(err)
				}
			}
		}
	}
}

func readFromSource(ctx context.Context, flagConfig *models.FlagConfig) []chan models.OplogEntry {
	select {
	case <-ctx.Done():
		return nil

	default:
		switch flagConfig.InputType {
		case constants.InputTypeJSON:
			channel, err := readFileContent(ctx, flagConfig.InputFilePath)
			if err != nil {
				log.Fatal(err)
			}
			return channel

		case constants.InputTypeMongoDB:
			if err := mongodb.InitializeMongoDb(); err != nil {
				log.Fatal(err)
			}
			dbCollections, err := mongodb.ListAllDbsAndCollections(ctx)
			if err != nil {
				log.Fatal(err)
			}

			oplogChannel := make([]chan models.OplogEntry, len(dbCollections))
			for i := range oplogChannel {
				oplogChannel[i] = make(chan models.OplogEntry)
			}

			for i, dbCollection := range dbCollections {
				go mongodb.WatchCollection(ctx, oplogChannel[i], dbCollection)
			}

			return oplogChannel
		}
	}

	return nil
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

// using round robin to consistently distribute oplogs to channels to be consumed.
func readFileContent(ctx context.Context, filePath string) ([]chan models.OplogEntry, error) {
	noOfChannels := 10
	channels := make([]chan models.OplogEntry, noOfChannels)
	for i := range channels {
		channels[i] = make(chan models.OplogEntry)
	}

	databaseCollections := make(map[string]int)
	var mu sync.Mutex

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	lastReadPoistion := constants.LastReadCheckpoint.GetFileCheckpoint()
	_, err = file.Seek(lastReadPoistion, io.SeekStart)
	if err != nil {
		log.Fatal(err)
	}

	decoder := json.NewDecoder(file)
	if lastReadPoistion == 0 {
		_, err = decoder.Token()
		if err != nil {
			log.Fatal(err)
		}
	}

	go func() {
		defer file.Close()
		defer func() {
			for i := range channels {
				close(channels[i])
			}
		}()
		for {
			if decoder.More() {
				var decodedData models.OplogEntry
				err = decoder.Decode(&decodedData)
				if err != nil {
					log.Fatal(err)
				}

				namespace := strings.Split(decodedData.Namespace, ".")
				if slices.Contains(constants.NotAllowedNameSpaceForFile, namespace[0]) ||
					!slices.Contains(constants.AllowedOperations, decodedData.Operation) {
					continue
				}

				mu.Lock()
				channelIndex, exists := databaseCollections[decodedData.Namespace]
				if !exists {
					channelIndex = hashNamespace(decodedData.Namespace, len(channels))
					databaseCollections[decodedData.Namespace] = channelIndex
				}
				mu.Unlock()

				select {
				case <-ctx.Done():
					return

				case channels[channelIndex] <- decodedData:
					readPosition, err := file.Seek(0, io.SeekCurrent)
					if err != nil {
						log.Fatal(err)
					}
					constants.LastReadCheckpoint.SetFileCheckpoint(readPosition)
				}
			}
		}
	}()

	return channels, nil
}

func hashNamespace(namespace string, lenght int) int {
	h := fnv.New32a()
	h.Write([]byte(namespace))
	return int(h.Sum32()) % lenght

}

func saveLastRead() error {
	gobFile, err := os.OpenFile("checkpoint.gob", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	err = constants.LastReadCheckpoint.EncodeToGob(gobFile)
	if err != nil {
		return err
	}

	err = gobFile.Close()
	if err != nil {
		return err
	}

	return nil
}

func restoreCheckpoint() error {
	gobFile, err := os.OpenFile("checkpoint.gob", os.O_RDWR, 0666)
	if os.IsNotExist(err) {
		return nil //if no checkpoint is created, then return nil
	}

	if err != nil {
		return err
	}
	defer gobFile.Close()

	fileInfo, err := gobFile.Stat()
	if err != nil {
		return err
	}

	if fileInfo.Size() != 0 {
		err = constants.LastReadCheckpoint.DecodeFromGob(gobFile)
		if err != nil {
			return err
		}
	}

	return nil
}
