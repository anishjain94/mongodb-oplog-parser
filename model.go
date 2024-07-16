package main

type Oplog struct {
	Operation EnumOperation `json:"op"` //operation
	Namespace string        `json:"ns"` //namespace -> database.table_name
	Object    StudentModel  `json:"o"`  //data for insertion/updation/deletion
}

type StudentModel struct {
	ID          string `json:"_id"`
	Name        string `json:"name"`
	RollNo      int    `json:"roll_no"`
	IsGraduated bool   `json:"is_graduated"`
	DateOfBirth string `json:"date_of_birth"`
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
