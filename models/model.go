package models

import (
	"github.com/anishjain94/mongo-oplog-to-sql/constants"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Oplog struct {
	Operation constants.EnumOperation `bson:"op" json:"op"`                     //operation
	Namespace string                  `bson:"ns" json:"ns"`                     //namespace -> database.table_name
	Timestamp primitive.Timestamp     `bson:"ts" json:"ts"`                     //timestamp
	Object    map[string]interface{}  `bson:"o" json:"o"`                       //data for insertion/updation/deletion
	Object2   map[string]interface{}  `bson:"o2,omitempty" json:"o2,omitempty"` //where clause data
}

type ForeignKeyRelation struct {
	ColumnName string
	Value      interface{}
}

type FlagConfig struct {
	InputType      constants.InputType
	OutputType     constants.OutputType
	InputFilePath  string
	OutputFilePath string
}

type ProcessOplog struct {
	Channel    <-chan Oplog
	FlagConfig *FlagConfig
}
