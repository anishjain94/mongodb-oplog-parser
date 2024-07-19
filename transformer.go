package main

import (
	"fmt"
	"log"
	"reflect"
	"slices"
	"strings"

	"github.com/google/uuid"
)

// TODO: write a unit test for this function.
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
	opLogNameSpace := strings.Split(opLog.Namespace, ".")
	schemaName := opLogNameSpace[0]
	parentTableName := opLogNameSpace[1]

	if !CreateSchemaQueryExists[schemaName] {
		createSchemaQuery := GetCreateSchemaQuery(schemaName)

		queries = append(queries, createSchemaQuery)
		CreateSchemaQueryExists[schemaName] = true
	}

	for key, value := range opLog.Object {
		tableName := opLog.Namespace + "_" + key
		foreignKeyName := parentTableName + "__id"

		foreignKeyValue := GetValueFromObject("_id", opLog.Object)

		switch v := value.(type) {
		case map[string]interface{}:
			query := handleQueryCreation(tableName, v, &ForeignKeyRelation{
				ColumnName: foreignKeyName,
				Value:      foreignKeyValue,
			})

			queries = append(queries, query...)
			delete(opLog.Object, key) //NOTE: deleting nested key so that this key does not appear in create table for document.

		case []interface{}:
			for _, item := range v {
				temp := item.(map[string]interface{})
				query := handleQueryCreation(tableName, temp, &ForeignKeyRelation{
					ColumnName: foreignKeyName,
					Value:      foreignKeyValue,
				})

				queries = append(queries, query...)
			}
			delete(opLog.Object, key) //NOTE: deleting nested key so that this key does not appear in create table for document.
		}
	}

	schemaQueries := handleQueryCreation(opLog.Namespace, opLog.Object, nil)

	queries = append(queries, schemaQueries...)
	return queries
}

func GetValueFromObject(key string, object map[string]interface{}) interface{} {
	if value, exists := object[key]; exists {
		return value
	}
	return nil
}

func handleQueryCreation(tableName string, data map[string]interface{}, foreignKeyRelation *ForeignKeyRelation) []string {
	var queries []string

	idColumnExists := slices.Contains(getKeys(data), "_id")
	// If id column does not exists that means that its a nested object. So we create _id and foreign key column
	if !idColumnExists {
		data["_id"] = uuid.New().String()
		if foreignKeyRelation != nil {
			data[foreignKeyRelation.ColumnName] = foreignKeyRelation.Value
		}
	}
	columnNames := getKeys(data)

	if !CreateTableQueryExists[tableName] {
		createTableQuery := GetCreateTableQuery(tableName, data)

		queries = append(queries, createTableQuery)
		CreateTableQueryExists[tableName] = true
		TableColumnName[tableName] = columnNames
	}

	alterQueries := GetCreateAlterQuery(tableName, data)
	queries = append(queries, alterQueries...)

	insertQuery := GetInsertTableQuery(tableName, data)
	queries = append(queries, insertQuery)

	return queries
}

func GetInsertTableQuery(tableName string, data map[string]interface{}) string {
	columns := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))
	placeholders := make([]string, 0, len(data))

	for key, value := range data {
		columns = append(columns, key)
		values = append(values, value)
		placeholders = append(placeholders, "?")
	}

	insertQuery := fmt.Sprintf(
		"INSERT INTO %s(%s) VALUES (%s);\n\n",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)
	// insert into tablename(id, name, phone) values (?, ?, ?);
	// TODO: can replace (?, ?, ?) with direct values instead of using placeholders.
	// TODO: read about prepared statements in sql
	insertQuery = populateValuesInQuery(insertQuery, values)
	return insertQuery
}

func GetCreateAlterQuery(tableName string, data map[string]interface{}) []string {
	existingColumns := TableColumnName[tableName]
	var alterStatements []string

	for columnName, value := range data {
		if exist := slices.Contains(existingColumns, columnName); !exist {
			dataType := getDataType(value)
			TableColumnName[tableName] = append(TableColumnName[tableName], columnName)

			alterStatements = append(alterStatements, fmt.Sprintf("ALTER TABLE %s ADD %s %s;\n\n", tableName, columnName, dataType))
		}
	}

	return alterStatements
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func GetCreateTableQuery(tableName string, data map[string]interface{}) string {
	var placeHolders []string
	for range len(data) {
		placeHolders = append(placeHolders, "?")
	}
	placeHolder := strings.Join(placeHolders, ",")

	for key, value := range data {
		dataType := getDataType(value) //TODO: why ints are getting decoded as floats.
		columnAndType := fmt.Sprintf("%s %s", key, dataType)

		if key == "_id" {
			columnAndType = fmt.Sprintf("%s %s", columnAndType, "PRIMARY KEY")
		}
		placeHolder = strings.Replace(placeHolder, "?", columnAndType, 1) //TODO: can use strings.buffer to create query.
	}

	createTableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(%s);\n\n",
		tableName,
		placeHolder,
	)

	return createTableQuery
}

func GetCreateSchemaQuery(schemaName string) string {
	return fmt.Sprintf("CREATE SCHEMA %s;\n\n", schemaName)
}

func getDataType(value interface{}) string {
	var dataType string
	switch value.(type) {
	case string:
		return "VARCHAR(255)"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "INTEGER"
	case float32, float64:
		if _, ok := value.(int); ok {
			return "INTEGER"
		}
		return "FLOAT"
	case bool:
		return "BOOLEAN"
	case nil:
		dataType = "NULL"
	}

	return dataType
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
