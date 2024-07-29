package transformer

import (
	"reflect"
	"testing"

	"github.com/anishjain94/mongo-oplog-to-sql/models"
)

var testOplogQuery = map[string]struct {
	models.Oplog
	Want []string
}{
	"insertSingle": {
		Oplog: models.Oplog{
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
	"insertSingleNewColumn": {
		Oplog: models.Oplog{
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
		Oplog: models.Oplog{
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
		Want: []string{"UPDATE test.student SET is_graduated = true, name = 'dummy_name' WHERE _id = '635b79e231d82a8ab1de863b';\n\n"},
	},
	"updateQuerySetNull": {
		Oplog: models.Oplog{
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
		Want: []string{"UPDATE test.student SET name = NULL, roll_no = NULL WHERE _id = '635b79e231d82a8ab1de863b';\n\n"},
	},
	"deleteQuery": {
		Oplog: models.Oplog{
			Operation: "d",
			Namespace: "test.student",
			Object: map[string]interface{}{
				"_id": "635b79e231d82a8ab1de863b",
			},
		},
		Want: []string{"DELETE FROM test.student WHERE _id = '635b79e231d82a8ab1de863b';\n\n"},
	},
	"nestedObject1": {
		Oplog: models.Oplog{
			Operation: "i",
			Namespace: "test.student",
			Object: map[string]interface{}{
				"_id":           "635b79e231d82a8ab1de863b",
				"name":          "Selena Miller",
				"roll_no":       51,
				"is_graduated":  false,
				"date_of_birth": "2000-01-30",
				"phone": map[string]interface{}{
					"personal": "7678456640",
					"work":     "8130097989",
				},
				"address": []map[string]interface{}{
					{
						"line1": "481 Harborsburgh",
						"zip":   "89799",
					},
					{
						"line1": "329 Flatside",
						"zip":   "80872",
					},
				},
			},
		},
	},
	"nestedObject2": {
		Oplog: models.Oplog{
			Operation: "i",
			Namespace: "test.student",
			Object: map[string]interface{}{
				"_id":           "635b79e231d82a8ab1de863b",
				"name":          "Selena Miller",
				"roll_no":       51,
				"is_graduated":  false,
				"date_of_birth": "2000-01-30",
				"phone": map[string]interface{}{
					"personal": "7678456640",
					"work":     "8130097989",
					"home":     "8989723",
				},
				"address": []map[string]interface{}{
					{
						"line1":   "481 Harborsburgh",
						"zip":     "89799",
						"pincode": "123",
					},
					{
						"line1": "329 Flatside",
						"zip":   "80872",
					},
				},
			},
		},
	},
}

func TestOplogGenereateQuery(t *testing.T) {
	for key, value := range testOplogQuery {
		t.Run(key, func(t *testing.T) {
			got := GetSqlQueries(value.Oplog)

			if len(value.Want) != 0 && !reflect.DeepEqual(got, value.Want) {
				t.Errorf("got : %s\nwant : %s", got, value.Want)
			}
		})
	}
}

func TestPopulateValuesInQuery(t *testing.T) {
	tests := map[string]struct {
		query    string
		values   []interface{}
		expected string
	}{
		"String replacement": {
			query:    "SELECT * FROM users WHERE name = ?",
			values:   []interface{}{"John"},
			expected: "SELECT * FROM users WHERE name = 'John'",
		},
		"Integer replacement": {
			query:    "SELECT * FROM users WHERE age > ?",
			values:   []interface{}{30},
			expected: "SELECT * FROM users WHERE age > 30",
		},
		"Float replacement": {
			query:    "SELECT * FROM products WHERE price < ?",
			values:   []interface{}{19.99},
			expected: "SELECT * FROM products WHERE price < 19.990000",
		},
		"Boolean replacement": {
			query:    "SELECT * FROM users WHERE is_active = ?",
			values:   []interface{}{true},
			expected: "SELECT * FROM users WHERE is_active = true",
		},
		"Null replacement": {
			query:    "SELECT * FROM users WHERE last_login = ?",
			values:   []interface{}{nil},
			expected: "SELECT * FROM users WHERE last_login = NULL",
		},
		"Multiple replacements": {
			query:    "INSERT INTO users (name, age, balance) VALUES (?, ?, ?)",
			values:   []interface{}{"Alice", 25, 1000.50},
			expected: "INSERT INTO users (name, age, balance) VALUES ('Alice', 25, 1000.500000)",
		},
		"String with single quotes": {
			query:    "SELECT * FROM users WHERE name = ?",
			values:   []interface{}{"O'Brien"},
			expected: "SELECT * FROM users WHERE name = 'O''Brien'",
		},
		"Different integer types": {
			query:    "SELECT * FROM data WHERE int8 = ? AND int16 = ? AND int32 = ? AND int64 = ?",
			values:   []interface{}{int8(8), int16(16), int32(32), int64(64)},
			expected: "SELECT * FROM data WHERE int8 = 8 AND int16 = 16 AND int32 = 32 AND int64 = 64",
		},
		"Different unsigned integer types": {
			query:    "SELECT * FROM data WHERE uint8 = ? AND uint16 = ? AND uint32 = ? AND uint64 = ?",
			values:   []interface{}{uint8(8), uint16(16), uint32(32), uint64(64)},
			expected: "SELECT * FROM data WHERE uint8 = 8 AND uint16 = 16 AND uint32 = 32 AND uint64 = 64",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := populateValuesInQuery(tt.query, tt.values)
			if result != tt.expected {
				t.Errorf("%s: populateValuesInQuery() = %v, want %v", name, result, tt.expected)
			}
		})
	}
}
