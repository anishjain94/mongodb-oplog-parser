package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// TODO: read about how and when is memory allocated to variables and global variables.
// TODO: handle race conditions for these global maps.
// TODO: think of a better name.

var CreateSchemaQueryExists map[string]bool
var CreateTableQueryExists map[string]bool
var TableColumnName map[string][]string

func init() {
	err := godotenv.Load("./.env")
	if err != nil {
		log.Fatal("unable to load " + ".env file")
	}
	CreateSchemaQueryExists = make(map[string]bool)
	CreateTableQueryExists = make(map[string]bool)
	TableColumnName = make(map[string][]string)
}

func main() {
	fileConfig := parseFlags()

	if *fileConfig.InputFilePath == "" {
		log.Fatalln("no input file")
	}

	var queries []string
	decodedData, err := readFile(*fileConfig)
	if err != nil {
		log.Fatal(err)
	}

	for _, logs := range decodedData {
		sqlQueries := GetSqlQueries(logs)
		if err != nil {
			log.Fatal(err)
		}
		queries = append(queries, sqlQueries...)
	}

	err = displayOutput(*fileConfig, queries)
	if err != nil {
		log.Fatal(err)
	}
}

func parseFlags() *FlagConfig {
	inputFilePath := flag.String("i", "example.json", "oplog json file path")
	outputFilePath := flag.String("o", "output.sql", "output file path")

	flag.Parse()
	fileConfig := FlagConfig{
		InputFilePath:  inputFilePath,
		OutputFilePath: outputFilePath,
	}

	return &fileConfig
}

func displayOutput(fileConfig FlagConfig, queries []string) error {
	file, err := os.OpenFile(*fileConfig.OutputFilePath, os.O_WRONLY|os.O_CREATE, 0666)
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

// TODO: handle permission problems, file does not exist, too many files opened.
func readFile(fileConfig FlagConfig) ([]Oplog, error) {
	file, err := os.Open(*fileConfig.InputFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var decodedData []Oplog

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&decodedData)
	if err != nil {
		return nil, err
	}

	return decodedData, nil
}

func GetSqlQueries(oplog Oplog) []string {
	var query []string

	switch oplog.Operation {
	case EnumOperationInsert:
		query = GetInsertQueryFromOplog(oplog)

	case EnumOperationUpdate:
		query = GetUpdateQueryFromOplog(oplog)

	case EnumOperationDelete:
		query = GetDeleteQueryFromOplog(oplog)

	}

	return query
}
