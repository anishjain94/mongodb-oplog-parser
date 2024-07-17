package main

type EnumOperation string

const (
	EnumOperationInsert EnumOperation = "i"
	EnumOperationUpdate EnumOperation = "u"
	EnumOperationDelete EnumOperation = "d"
)

var ShouldCreateSchemaQuery = make(map[string]bool)
var ShouldCreateTableQuery = make(map[string]bool)
var TableColumnName = make(map[string][]string)
