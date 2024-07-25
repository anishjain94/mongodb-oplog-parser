package transformer

import (
	"fmt"
	"slices"
	"strings"

	"github.com/anishjain94/mongo-oplog-to-sql/constants"
	"github.com/anishjain94/mongo-oplog-to-sql/models"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func GetSqlQueries(oplog models.Oplog) []string {
	var query []string

	switch oplog.Operation {
	case constants.EnumOperationInsert:
		query = GetInsertQueryFromOplog(oplog)

	case constants.EnumOperationUpdate:
		query = append(query, GetUpdateQueryFromOplog(oplog))

	case constants.EnumOperationDelete:
		query = append(query, GetDeleteQueryFromOplog(oplog))
	}

	return query
}

// TODO: explore use of generics.
func populateValuesInQuery(query string, values []interface{}) string {
	for _, v := range values {
		replace := "?"
		switch v := v.(type) {
		case string:
			replace = fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64: //TODO: better way to do this.
			replace = fmt.Sprintf("%d", v)
		case float32, float64:
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

func GetInsertQueryFromOplog(opLog models.Oplog) []string {
	var queries []string
	opLogNameSpace := strings.Split(opLog.Namespace, ".")
	schemaName := opLogNameSpace[0]
	parentTableName := opLogNameSpace[1]

	globalVariables := constants.GetGlobalVariables()

	if _, exists := globalVariables.CreateSchemaQuery.Get(schemaName); !exists {
		createSchemaQuery := GetCreateSchemaQuery(schemaName)

		queries = append(queries, createSchemaQuery)
		globalVariables.CreateSchemaQuery.Set(schemaName, true)
	}

	for key, value := range opLog.Object {
		tableName := opLog.Namespace + "_" + key
		foreignKeyName := parentTableName + "__id"
		foreignKeyValue := GetValueFromObject("_id", opLog.Object)
		switch data := value.(type) {
		case map[string]interface{}:
			query := handleQueryCreation(tableName, data, &models.ForeignKeyRelation{
				ColumnName: foreignKeyName,
				Value:      foreignKeyValue,
			})

			queries = append(queries, query...)
			delete(opLog.Object, key) //NOTE: deleting nested key so that this key does not appear in create table for document.

		case []interface{}:
			for _, item := range data {
				temp := item.(map[string]interface{}) //type assertion
				query := handleQueryCreation(tableName, temp, &models.ForeignKeyRelation{
					ColumnName: foreignKeyName,
					Value:      foreignKeyValue,
				})

				queries = append(queries, query...)
			}
			delete(opLog.Object, key) //NOTE: deleting nested key so that this key does not appear in create table for document.

		case bson.A:
			temp := []interface{}(data) //type conversion
			for _, item := range temp {
				item := item.(map[string]interface{})
				query := handleQueryCreation(tableName, item, &models.ForeignKeyRelation{
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

func handleQueryCreation(tableName string, data map[string]interface{}, foreignKeyRelation *models.ForeignKeyRelation) []string {
	var queries []string
	globalVariables := constants.GetGlobalVariables()

	idColumnExists := slices.Contains(getKeys(data), "_id")
	// If id column does not exists that means that its a nested object. So we create _id and foreign key column
	if !idColumnExists {
		data["_id"] = uuid.New().String()
		if foreignKeyRelation != nil {
			data[foreignKeyRelation.ColumnName] = foreignKeyRelation.Value
		}
	}
	columnNames := getKeys(data)

	if _, exists := globalVariables.CreateTableQuery.Get(tableName); !exists {
		createTableQuery := GetCreateTableQuery(tableName, data)

		queries = append(queries, createTableQuery)
		globalVariables.CreateTableQuery.Set(tableName, true)
		globalVariables.TableColumnName.Set(tableName, columnNames)
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
	// TODO: can replace (?, ?, ?) with direct values instead of using placeholders.
	insertQuery = populateValuesInQuery(insertQuery, values)
	return insertQuery
}

func GetCreateAlterQuery(tableName string, data map[string]interface{}) []string {
	globalVariables := constants.GetGlobalVariables()

	existingColumns, _ := globalVariables.TableColumnName.Get(tableName)
	var alterStatements []string

	for columnName, value := range data {
		if exist := slices.Contains(existingColumns, columnName); !exist {
			dataType := getDataType(value)

			columns, _ := globalVariables.TableColumnName.Get(tableName)
			columns = append(columns, columnName)
			globalVariables.TableColumnName.Set(tableName, columns)

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
		dataType := getDataType(value) //TODO: why ints are getting converted as floats.
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
	return fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s;\n\n", schemaName)
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

func GetUpdateQueryFromOplog(opLog models.Oplog) string {
	dataToUpdate := make(map[string]interface{})
	dataToSetNull := make(map[string]interface{})

	whereClause := opLog.Object2
	if diff, ok := opLog.Object["diff"].(map[string]interface{}); ok {
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
		values = append(values, nil) //TODO: change to zero value for that datatype. check if can be done.
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

	return updateQuery
}

func GetDeleteQueryFromOplog(opLog models.Oplog) string {
	where := make([]string, 0, len(opLog.Object))
	values := make([]interface{}, 0, len(opLog.Object))

	// where clause
	for key, value := range opLog.Object {
		where = append(where, fmt.Sprintf("%s = ?", key))
		values = append(values, value)
	}

	deleteQuery := fmt.Sprintf("DELETE FROM %s WHERE %s;\n\n",
		opLog.Namespace,
		strings.Join(where, " and "),
	)

	deleteQuery = populateValuesInQuery(deleteQuery, values)

	return deleteQuery
}
