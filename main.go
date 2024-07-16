package main

import (
	"encoding/json"
	"fmt"
	"log"
)

func main() {
	var query string

	for _, oplog := range exampleOpLogs {
		var opLog Oplog
		bytesData, err := json.Marshal(oplog)

		if err != nil {
			log.Fatalf(err.Error())
		}

		err = json.Unmarshal(bytesData, &opLog)
		if err != nil {
			log.Fatalf(err.Error())
		}

		switch opLog.Operation {
		case EnumOperationInsert:
			query = GetInsertQueryFromOplogUsingMap(opLog)

		case EnumOperationUpdate:
			query = GetInsertQueryFromOplogUsingMap(opLog)

		case EnumOperationDelete:
			query = GetInsertQueryFromOplogUsingMap(opLog)

		}
		fmt.Println(query)
	}
}
