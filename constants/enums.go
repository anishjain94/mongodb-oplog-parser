package constants

type EnumOperation string

const (
	EnumOperationInsert EnumOperation = "i"
	EnumOperationUpdate EnumOperation = "u"
	EnumOperationDelete EnumOperation = "d"
)

type InputType string
type OutputType string

const (
	InputTypeJSON    InputType  = "json"
	InputTypeMongoDB InputType  = "mongodb"
	OutputTypeSQL    OutputType = "sql"
	OutputTypeDB     OutputType = "db"
)
