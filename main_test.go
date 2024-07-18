package main

import "testing"

func TestMain(t *testing.T) {
	var queries []string

	inputFile := "example.json"
	outputFile := "output.sql"

	config := FlagConfig{
		InputFilePath:  &inputFile,
		OutputFilePath: &outputFile,
	}
	decodedData := readFile(config)

	for _, logs := range decodedData {
		queriesToAppend, err := transformHandler(logs)
		if err != nil {
			t.Error(err)
		}
		queries = append(queries, queriesToAppend...)
	}

	displayOutput(config, queries)
}
