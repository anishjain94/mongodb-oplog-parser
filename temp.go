package main

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

func GetInsertQueryFromOplogTemp(opLog Oplog) string {
	objectMap := make(map[string]interface{})
	value := reflect.ValueOf(opLog.Object)

	// Check if it's a struct
	if value.Kind() != reflect.Struct {
		log.Panicf("Object field is not a struct")
	}

	typeOf := value.Type()

	for i := 0; i < value.NumField(); i++ {
		field := typeOf.Field(i)
		fieldValue := value.Field(i)

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = field.Name
		}

		objectMap[jsonTag] = fieldValue.Interface()
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
