package main

type EnumOperation string

const (
	EnumOperationInsert EnumOperation = "i"
	EnumOperationUpdate EnumOperation = "u"
	EnumOperationDelete EnumOperation = "d"
)

var CreateSchemaQueryExists = make(map[string]bool)
var CreateTableQueryExists = make(map[string]bool)
var TableColumnName = make(map[string][]string)
