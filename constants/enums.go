package constants

type EnumOperation string

const (
	EnumOperationInsert EnumOperation = "i"
	EnumOperationUpdate EnumOperation = "u"
	EnumOperationDelete EnumOperation = "d"
)

var AllowedOperations = []EnumOperation{
	EnumOperationInsert,
	EnumOperationUpdate,
	EnumOperationDelete,
}

var NotAllowedNameSpaceForFile = []string{
	"admin",
	"config",
	"local",
}

type InputType string
type OutputType string

const (
	InputTypeJSON    InputType  = "json"
	InputTypeMongoDB InputType  = "mongodb"
	OutputTypeSQL    OutputType = "sql"
	OutputTypeDB     OutputType = "db"
)
