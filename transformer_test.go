package main

import (
	"fmt"
	"testing"
)

// kind is underlying datatype
// type is the userDefined data type.

var testOplogQuery = map[string]struct {
	Oplog
	want string
}{
	"insertSingle": {
		Oplog: Oplog{
			Operation: "i",
			Namespace: "test.student",
			Object: map[string]interface{}{
				"_id":           "635b79e231d82a8ab1de863b",
				"name":          "Selena Miller",
				"roll_no":       51,
				"is_graduated":  false,
				"date_of_birth": "2000-01-30",
			},
		},
		// TODO: ask mohit on how to write test cases in this case..
		// want: "INSERT INTO test.students(_id, name, roll_no, is_graduated, date_of_birth) VALUES ('635b79e231d82a8ab1de863b', 'Selena Miller', 51, false, '2000-01-30')",
	},
}

func TestOplogInsertQuery(t *testing.T) {
	for key, value := range testOplogQuery {
		t.Run(key, func(t *testing.T) {
			insertQuery := GetInsertQueryFromOplogUsingMap(value.Oplog)
			fmt.Println(insertQuery)
			if insertQuery == "" {
				t.Errorf("got : %s \nwant:%s", insertQuery, value.want)
			}
		})
	}

}
