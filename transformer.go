package main

import (
	"fmt"
	"log"
	"reflect"
	"slices"
	"strings"

	"github.com/google/uuid"
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

func GetInsertQueryFromOplog(opLog Oplog) ([]string, error) {
	if reflect.ValueOf(opLog.Object).Kind() != reflect.Map {
		return nil, fmt.Errorf("object field is not a struct")
	}

	var queries []string

	if !CreateSchemaQueryExists[opLog.Namespace] {
		createSchemaQuery, err := GetCreateSchemaQuery(opLog.Namespace)
		if err != nil {
			return nil, err
		}
		queries = append(queries, createSchemaQuery)
		CreateSchemaQueryExists[opLog.Namespace] = true
	}

	for key, value := range opLog.Object {
		tableName := opLog.Namespace + "_" + key
		nameSpace := strings.Split(opLog.Namespace, ".")
		foreignKeyName := nameSpace[1] + "__id"

		foreignKeyValue, err := GetValueFromObject("_id", opLog.Object)
		if err != nil {
			return nil, err
		}

		switch v := value.(type) {
		case map[string]interface{}:
			query, err := handleMapValue(tableName, v, &ForeignKeyRelation{
				ColumnName: foreignKeyName,
				Value:      foreignKeyValue,
			})
			if err != nil {
				return nil, err
			}
			queries = append(queries, query...)
			delete(opLog.Object, key) //NOTE: deleting nested key so that this key does not appear in create table for document.

		case []interface{}:
			for _, item := range v {
				temp := item.(map[string]interface{})
				query, err := handleMapValue(tableName, temp, &ForeignKeyRelation{
					ColumnName: foreignKeyName,
					Value:      foreignKeyValue,
				})
				if err != nil {
					return nil, err
				}
				queries = append(queries, query...)
			}
			delete(opLog.Object, key) //NOTE: deleting nested key so that this key does not appear in create table for document.
		}
	}

	schemaQueries, err := handleMapValue(opLog.Namespace, opLog.Object, nil)
	if err != nil {
		return nil, err
	}

	queries = append(queries, schemaQueries...)
	return queries, nil
}

func GetValueFromObject(key string, object map[string]interface{}) (interface{}, error) {
	if value, exists := object[key]; exists {
		return value, nil
	}
	return nil, fmt.Errorf("%s not found", key)
}

func handleMapValue(key string, mapValue map[string]interface{}, foreignKeyRelation *ForeignKeyRelation) ([]string, error) {
	var queries []string
	columnNames := getKeys(mapValue)

	idColumnExists := slices.Contains(columnNames, "_id")

	// If id column does not exists that means that its a nested object. So we create _id and foreign key column
	if !idColumnExists {
		mapValue["_id"] = uuid.New().String()
		if foreignKeyRelation != nil {
			mapValue[foreignKeyRelation.ColumnName] = foreignKeyRelation.Value
		} else {
			return nil, fmt.Errorf("no foreign key id found")
		}
	}

	if !CreateTableQueryExists[key] {
		createTableQuery, err := GetCreateTableQuery(key, mapValue)
		if err != nil {
			return nil, err
		}
		queries = append(queries, createTableQuery)
		CreateTableQueryExists[key] = true
		TableColumnName[key] = columnNames
	}

	if len(columnNames) != len(TableColumnName[key]) {
		alterQuery, err := GetCreateAlterQuery(key, mapValue)
		if err != nil {
			return nil, err
		}
		queries = append(queries, alterQuery)
	}

	insertQuery, err := GetInsertTableQuery(key, mapValue)
	if err != nil {
		return nil, err
	}
	queries = append(queries, insertQuery)

	return queries, nil
}

// TODO: error
func GetInsertTableQuery(namespace string, objectMap map[string]interface{}) (string, error) {
	columns := make([]string, 0, len(objectMap))
	values := make([]interface{}, 0, len(objectMap))
	var placeholders []string

	for key, value := range objectMap {
		columns = append(columns, key)
		values = append(values, value)
		placeholders = append(placeholders, "?")
	}

	insertQuery := fmt.Sprintf(
		"INSERT INTO %s(%s) VALUES (%s);\n\n",
		namespace,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)
	insertQuery = populateValuesInQuery(insertQuery, values)
	return insertQuery, nil
}

func GetCreateAlterQuery(namespace string, objectMap map[string]interface{}) (string, error) {
	existingColumns := TableColumnName[namespace]

	for columnName, value := range objectMap {
		if exist := slices.Contains(existingColumns, columnName); !exist {
			dataType := getDataType(value)
			TableColumnName[namespace] = append(TableColumnName[namespace], columnName)

			return fmt.Sprintf("ALTER TABLE %s ADD %s %s;\n\n", namespace, columnName, dataType), nil
		}
	}

	return "", fmt.Errorf("cannot create alter table")
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func GetCreateTableQuery(namespace string, objectMap map[string]interface{}) (string, error) {
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

	createTableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(%s);\n\n",
		namespace,
		placeHolder,
	)

	return createTableQuery, nil
}

func GetCreateSchemaQuery(namespace string) (string, error) {
	namespaces := strings.Split(namespace, ".")

	return fmt.Sprintf("CREATE SCHEMA %s;\n\n", namespaces[0]), nil
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

	whereClause := opLog.Object2

	if diff, ok := objectMap["diff"].(map[string]interface{}); ok {
		if u, ok := diff["u"].(map[string]interface{}); ok {
			dataToUpdate = u
		} else if d, ok := diff["d"].(map[string]interface{}); ok {
			dataToSetNull = d
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

	updateQuery := fmt.Sprintf("UPDATE %s SET %s WHERE %s;\n\n",
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

	updateQuery := fmt.Sprintf("DELETE FROM %s WHERE %s;\n\n",
		opLog.Namespace,
		strings.Join(where, " and "),
	)

	updateQuery = populateValuesInQuery(updateQuery, values)

	return []string{updateQuery}
}
