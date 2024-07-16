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

func GetInsertQueryFromOplog(opLog Oplog) string {
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

func GetUpdateQueryFromOplog(opLog Oplog) string {
	objectMap := make(map[string]interface{})
	value := reflect.ValueOf(opLog.Object)
	dataToUpdate := make(map[string]interface{})
	dataToSetNull := make(map[string]interface{})
	whereClause := make(map[string]interface{})

	if value.Kind() != reflect.Map {
		log.Panicf("Object field is not a struct")
	}

	for key, value := range opLog.Object {
		objectMap[key] = value
	}

	whereClause = getDataFromInterface(objectMap["o2"].(map[string]interface{}))

	if diff, ok := objectMap["diff"].(map[string]interface{}); ok {
		if u, ok := diff["u"].(map[string]interface{}); ok {
			dataToUpdate = getDataFromInterface(u)
		} else if d, ok := diff["d"].(map[string]interface{}); ok {
			dataToSetNull = getDataFromInterface(d)
		}
	}

	setClause := make([]string, 0, len(dataToUpdate))
	values := make([]interface{}, 0, len(dataToUpdate))
	where := make([]string, 0, len(whereClause))

	// for updation
	for key, value := range dataToUpdate {
		setClause = append(setClause, fmt.Sprintf("%s = ?", key))
		values = append(values, value)
	}

	// for setting key as null
	for key := range dataToSetNull {
		setClause = append(setClause, fmt.Sprintf("%s = ?", key))
		values = append(values, nil)
	}

	// for where clause
	for key, value := range whereClause {
		where = append(where, fmt.Sprintf("%s = ?", key))
		values = append(values, value)
	}

	updateQuery := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		opLog.Namespace,
		strings.Join(setClause, ", "),
		strings.Join(where, " and "),
	)

	updateQuery = populateValuesInQuery(updateQuery, values)

	return updateQuery
}

func GetDeleteQueryFromOplog(opLog Oplog) string {
	objectMap := make(map[string]interface{})
	value := reflect.ValueOf(opLog.Object)

	if value.Kind() != reflect.Map {
		log.Panicf("Object field is not a struct")
	}

	for key, value := range opLog.Object {
		objectMap[key] = value
	}

	where := make([]string, 0, len(objectMap))
	values := make([]interface{}, 0, len(objectMap))

	// for where clause
	for key, value := range objectMap {
		where = append(where, fmt.Sprintf("%s = ?", key))
		values = append(values, value)
	}

	updateQuery := fmt.Sprintf("DELETE FROM %s WHERE %s",
		opLog.Namespace,
		strings.Join(where, " and "),
	)

	updateQuery = populateValuesInQuery(updateQuery, values)

	return updateQuery
}

func getDataFromInterface(data map[string]interface{}) map[string]interface{} {
	objectMap := make(map[string]interface{})
	for key, value := range data {
		objectMap[key] = value
	}

	return objectMap
}
