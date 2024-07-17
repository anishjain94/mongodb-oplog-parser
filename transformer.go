package main

import (
	"fmt"
	"log"
	"reflect"
	"slices"
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

func GetInsertQueryFromOplog(opLog Oplog) []string {
	var queries []string
	objectMap := make(map[string]interface{})
	var columnsNames []string

	// TODO: how do i handle error here.
	if reflect.ValueOf(opLog.Object).Kind() != reflect.Map {
		log.Panicf("Object field is not a struct")
	}

	// TODO: remove this.
	for key, value := range opLog.Object {
		objectMap[key] = value
		columnsNames = append(columnsNames, key)
	}

	if _, exists := ShouldCreateSchemaQuery[opLog.Namespace]; !exists {
		createSchemaQuery := GetCreateSchemaQuery(opLog.Namespace)
		queries = append(queries, createSchemaQuery)
		ShouldCreateSchemaQuery[opLog.Namespace] = true
	}

	if _, exists := ShouldCreateTableQuery[opLog.Namespace]; !exists {
		createTableQuery := GetCreateTableQuery(opLog.Namespace, objectMap)
		queries = append(queries, createTableQuery)
		ShouldCreateTableQuery[opLog.Namespace] = true
		TableColumnName[opLog.Namespace] = columnsNames
	}

	if len(columnsNames) != len(TableColumnName[opLog.Namespace]) {
		alterQuery := GetCreateAlterQuery(opLog.Namespace, objectMap)
		queries = append(queries, alterQuery)
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
	queries = append(queries, insertQuery)
	return queries
}

func GetCreateAlterQuery(namespace string, objectMap map[string]interface{}) string {
	existingColumns := TableColumnName[namespace]

	for key, value := range objectMap {
		if exist := slices.Contains(existingColumns, key); !exist {
			dataType := getDataType(value)
			return fmt.Sprintf("ALTER TABLE %s ADD %s %s;", namespace, key, dataType)
		}
	}

	return ""
}

func GetCreateTableQuery(namespace string, objectMap map[string]interface{}) string {
	placeholders := make([]string, 0, len(objectMap))
	for range len(objectMap) {
		placeholders = append(placeholders, "?")
	}
	placeHolder := strings.Join(placeholders, ", ")

	for key, value := range objectMap {
		dataType := getDataType(value)
		columnAndType := fmt.Sprintf("%s %s", key, dataType)

		if key == "_id" {
			columnAndType = fmt.Sprintf("%s %s", columnAndType, "PRIMARY KEY")
		}
		placeHolder = strings.Replace(placeHolder, "?", columnAndType, 1)
	}

	createTableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(%s);\n",
		namespace,
		placeHolder,
	)

	return createTableQuery
}

func GetCreateSchemaQuery(namespace string) string {
	namespaces := strings.Split(namespace, ".")

	return fmt.Sprintf("CREATE SCHEMA %s;\n", namespaces[0])
}

func getDataType(value interface{}) string {
	var replace string
	switch value.(type) {
	case string:
		replace = "VARCHAR(255)"
	case int, int64:
		replace = "INTEGER"
	case float64:
		replace = "FLOAT"
	case bool:
		replace = "BOOLEAN"
	case nil:
		replace = "NULL"
	}

	return replace
}

func GetUpdateQueryFromOplog(opLog Oplog) []string {
	objectMap := make(map[string]interface{})
	value := reflect.ValueOf(opLog.Object)
	dataToUpdate := make(map[string]interface{})
	dataToSetNull := make(map[string]interface{})

	if value.Kind() != reflect.Map {
		log.Panicf("Object field is not a struct")
	}

	for key, value := range opLog.Object {
		objectMap[key] = value
	}

	whereClause := getDataFromInterface(opLog.Object2)

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

	return []string{updateQuery}
}

func GetDeleteQueryFromOplog(opLog Oplog) []string {
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

	return []string{updateQuery}
}

func getDataFromInterface(data map[string]interface{}) map[string]interface{} {
	objectMap := make(map[string]interface{})
	for key, value := range data {
		objectMap[key] = value
	}

	return objectMap
}
