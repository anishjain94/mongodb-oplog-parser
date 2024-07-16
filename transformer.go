package main

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

// kind is underlying datatype
// type is the userDefined data type.

func GetInsertQueryFromOplog(opLog Oplog) string {
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

	insertQuery = InterpolateQuery(insertQuery, values)

	return insertQuery

}

func InterpolateQuery(query string, values []interface{}) string {
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
