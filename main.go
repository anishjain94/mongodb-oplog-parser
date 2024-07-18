package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"log"
	"os"
)

func main() {
	fileConfig := parseFlags()

	if *fileConfig.InputFilePath == "" {
		log.Fatalln("no input file")
	}

	var queries []string
	decodedData := readFile(fileConfig)

	for _, logs := range decodedData {
		queriesToAppend, err := transformHandler(logs)
		if err != nil {
			log.Fatal(err)
		}
		queries = append(queries, queriesToAppend...)
	}

	err := displayOutput(fileConfig, queries)
	if err != nil {
		log.Fatal(err)
	}
}

func parseFlags() FlagConfig {
	inputFilePath := flag.String("i", "", "oplog json file path")
	outputFilePath := flag.String("o", "output.sql", "output file path")

	flag.Parse()
	fileConfig := FlagConfig{
		InputFilePath:  inputFilePath,
		OutputFilePath: outputFilePath,
	}

	return fileConfig
}

func displayOutput(fileConfig FlagConfig, queries []string) error {
	file, err := os.OpenFile(*fileConfig.OutputFilePath, os.O_WRONLY|os.O_CREATE, 0666)
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

func readFile(fileConfig FlagConfig) []Oplog {
	file, err := os.Open(*fileConfig.InputFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var decodedData []Oplog

	decoder := json.NewDecoder(file)
	decoder.Decode(&decodedData)
	return decodedData
}

func transformHandler(oplog Oplog) ([]string, error) {
	var query []string
	var err error

	switch oplog.Operation {
	case EnumOperationInsert:
		query, err = GetInsertQueryFromOplog(oplog)

	case EnumOperationUpdate:
		query = GetUpdateQueryFromOplog(oplog)

	case EnumOperationDelete:
		query = GetDeleteQueryFromOplog(oplog)

	}

	if err != nil {
		return nil, err
	}

	return query, nil
}
