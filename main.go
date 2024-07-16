package main

import (
	"encoding/json"
	"fmt"
	"log"
)

func main() {

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

		insertQuery := GetInsertQueryFromOplog(opLog)
		fmt.Println(insertQuery)
	}
}
