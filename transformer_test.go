package main

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"
)

var testOplogQuery = map[string]struct {
	Oplog
	Want string
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
				"o2": map[string]interface{}{
					"_id": "635b79e231d82a8ab1de863b",
				},
			},
		},
		Want: "UPDATE test.student SET is_graduated = true, name = 'dummy_name' WHERE _id = '635b79e231d82a8ab1de863b'",
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
				"o2": map[string]interface{}{
					"_id": "635b79e231d82a8ab1de863b",
				},
			},
		},
		Want: "UPDATE test.student SET name = NULL, roll_no = NULL WHERE _id = '635b79e231d82a8ab1de863b'",
	},
	"deleteQuery": {
		Oplog: Oplog{
			Operation: "d",
			Namespace: "test.student",
			Object: map[string]interface{}{
				"_id": "635b79e231d82a8ab1de863b",
			},
		},
		Want: "DELETE FROM test.student WHERE _id = '635b79e231d82a8ab1de863b'",
	},
}

func TestOplogGenereateQuery(t *testing.T) {
	for key, value := range testOplogQuery {
		t.Run(key, func(t *testing.T) {
			got := transformHandler(value.Oplog)
			if (value.Want != "" && got != value.Want) || got == "" {
				t.Errorf("got : %s", got)
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
