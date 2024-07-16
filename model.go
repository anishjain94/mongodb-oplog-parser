package main

type Oplog struct {
	Operation EnumOperation          `json:"op"` //operation
	Namespace string                 `json:"ns"` //namespace -> database.table_name
	Object    map[string]interface{} `json:"o"`  //data for insertion/updation/deletion

	// For updation and deletion
	Object2 map[string]interface{} `json:"o2,omitempty"`
}
