package constants

func init() {
	CreateSchemaQueryExists = make(map[string]bool)
	CreateTableQueryExists = make(map[string]bool)
	TableColumnName = make(map[string][]string)
}
