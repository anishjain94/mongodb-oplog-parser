package main

type Oplog struct {
	Operation EnumOperation          `bson:"op" json:"op"` //operation
	Namespace string                 `bson:"ns" json:"ns"` //namespace -> database.table_name
	Object    map[string]interface{} `bson:"o" json:"o"`   //data for insertion/updation/deletion

	// For updation and deletion
	Object2 map[string]interface{} `bson:"o2,omitempty" json:"o2,omitempty"`
}

type ForeignKeyRelation struct {
	ColumnName string
	Value      interface{}
}

// TODO: ask chinmay on how to test when ordering of columns is inconsistent.
