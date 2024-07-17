package main

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"slices"
	"testing"
)

var testOplogQuery = map[string]struct {
	Oplog
	Want []string
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
		// Want: []string{
		// 	"CREATE SCHEMA test;",
		// 	"CREATE TABLE IF NOT EXISTS test.student(_id VARCHAR(255) PRIMARY KEY, name VARCHAR(255), roll_no INTEGER, is_graduated BOOLEAN, date_of_birth VARCHAR(255));",
		// },
	},
	"insertSingleNewColumn": {
		Oplog: Oplog{
			Operation: "i",
			Namespace: "test.student",
			Object: map[string]interface{}{
				"_id":           "635b79e231d82a8ab1de863b",
				"name":          "Selena Miller",
				"roll_no":       51,
				"is_graduated":  false,
				"date_of_birth": "2000-01-30",
				"phone":         "+91-81254966457",
			},
		},
	},
	"updateQuery": {
		Oplog: Oplog{
			Operation: "u",
			Namespace: "test.student",
			Object: map[string]interface{}{
				"$v": 2,
				"diff": map[string]interface{}{
					"u": map[string]interface{}{
						"is_graduated": true,
						"name":         "dummy_name",
					},
				},
			},
			Object2: map[string]interface{}{
				"_id": "635b79e231d82a8ab1de863b",
			},
		},
		Want: []string{"UPDATE test.student SET is_graduated = true, name = 'dummy_name' WHERE _id = '635b79e231d82a8ab1de863b'"},
	},
	"updateQuerySetNull": {
		Oplog: Oplog{
			Operation: "u",
			Namespace: "test.student",
			Object: map[string]interface{}{
				"$v": 2,
				"diff": map[string]interface{}{
					"d": map[string]interface{}{
						"roll_no": false,
						"name":    nil,
					},
				},
			},
			Object2: map[string]interface{}{
				"_id": "635b79e231d82a8ab1de863b",
			},
		},
		Want: []string{"UPDATE test.student SET name = NULL, roll_no = NULL WHERE _id = '635b79e231d82a8ab1de863b'"},
	},
	"deleteQuery": {
		Oplog: Oplog{
			Operation: "d",
			Namespace: "test.student",
			Object: map[string]interface{}{
				"_id": "635b79e231d82a8ab1de863b",
			},
		},
		Want: []string{"DELETE FROM test.student WHERE _id = '635b79e231d82a8ab1de863b'"},
	},
}

func TestOplogGenereateQuery(t *testing.T) {
	for key, value := range testOplogQuery {
		t.Run(key, func(t *testing.T) {
			got := transformHandler(value.Oplog)

			slices.Sort(got)
			slices.Sort(value.Want)

			if len(value.Want) != 0 && !reflect.DeepEqual(got, value.Want) {
				t.Errorf("got : %s\nwant : %s", got, value.Want)
			}
		})
	}
}

func TestOpLogGeneric(t *testing.T) {
	jsonStr := `{
		"op": "u",
		"ns": "test.student",
		"o": {
		   "$v": 2,
		   "diff": {
			  "u": {
				 "is_graduated": true
			  }
		   }
		},
		 "o2": {
		   "_id": "635b79e231d82a8ab1de863b"
		}
	 }`
	var oplog Oplog
	err := json.Unmarshal([]byte(jsonStr), &oplog)
	if err != nil {
		log.Fatal(err.Error())
	}

	query := transformHandler(oplog)
	fmt.Println(query)
}
