package main

type Oplog struct {
	Operation EnumOperation          `json:"op"` //operation
	Namespace string                 `json:"ns"` //namespace -> database.table_name
	Object    map[string]interface{} `json:"o"`  //data for insertion/updation/deletion
}

var exampleOpLogs = []map[string]interface{}{
	{
		"op": "i",
		"ns": "test.student",
		"o": map[string]interface{}{
			"_id":           "635b79e231d82a8ab1de863b",
			"name":          "Selena Miller",
			"roll_no":       51,
			"is_graduated":  false,
			"date_of_birth": "2000-01-30",
		},
	},
}
