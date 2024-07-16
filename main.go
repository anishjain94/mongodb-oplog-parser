package main

import (
	"fmt"
)

func transformHandler(oplog Oplog) string {
	var query string

	switch oplog.Operation {
	case EnumOperationInsert:
		query = GetInsertQueryFromOplog(oplog)

	case EnumOperationUpdate:
		query = GetUpdateQueryFromOplog(oplog)

	case EnumOperationDelete:
		query = GetDeleteQueryFromOplog(oplog)

	}
	fmt.Println(query)
	return query
}
