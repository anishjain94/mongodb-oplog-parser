package main

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

func populateValuesInQuery(query string, values []interface{}) string {
	for _, v := range values {
		replace := "?"
		switch v := v.(type) {
		case string:
			replace = fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
		case int, int64:
			replace = fmt.Sprintf("%d", v)
		case float64:
			replace = fmt.Sprintf("%f", v)
		case bool:
			replace = fmt.Sprintf("%t", v)
		case nil:
			replace = "NULL"
		}
		query = strings.Replace(query, "?", replace, 1)
	}
	return query
}

func GetInsertQueryFromOplogUsingMap(opLog Oplog) string {
	objectMap := make(map[string]interface{})
	value := reflect.ValueOf(opLog.Object)

	if value.Kind() != reflect.Map {
		log.Panicf("Object field is not a struct")
	}

	for key, value := range opLog.Object {
		objectMap[key] = value
	}

	columns := make([]string, 0, len(objectMap))
	values := make([]interface{}, 0, len(objectMap))
	var placeholders []string

	for key, value := range objectMap {
		columns = append(columns, key)
		values = append(values, value)
		placeholders = append(placeholders, "?")
	}

	insertQuery := fmt.Sprintf(
		"INSERT INTO %s(%s) VALUES (%s)",
		opLog.Namespace,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	insertQuery = populateValuesInQuery(insertQuery, values)

	return insertQuery
}

