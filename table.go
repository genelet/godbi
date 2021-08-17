package godbi

import (
	"encoding/json"
	"regexp"
	"strings"
)

type Table struct {
	CurrentTable  string   `json:"table" hcl:"table"`
	CurrentKeys   []string `json:"pks,omitempty" hcl:"pks,optional"`
	CurrentIDAuto string   `json:"id_auto,omitempty" hcl:"id_auto,optional"`
	Fks           []string `json:"fks,omitempty" hcl:"fks,optional"`
}

func filterPars(selectPars map[string][2]string, fields []string) map[string][2]string {
	if fields == nil || selectPars == nil {
		return nil
	}

	labels := make(map[string][2]string)
	for key, val := range selectPars {
		if grep(fields, key) {
			labels[key] = val
		}
	}
	return labels
}

func newTable(content []byte) (*Table, error) {
	var parsed *Table
	if err := json.Unmarshal(content, &parsed); err != nil {
		return nil, err
	}

	parsed.topicsHashPars = generalHashPars(parsed.TopicsHash, parsed.TopicsPars, nil)
	parsed.editHashPars   = generalHashPars(parsed.EditHash, parsed.EditPars, nil)
	parsed.DefaultSelectNames()
	return parsed, nil
}

func (self *Table) DefaultSelectNames() {
	if self.Sortby == "" {
		self.Sortby = "sortby"
	}
	if self.Sortreverse == "" {
		self.Sortreverse = "sortreverse"
	}
	if self.Pageno == "" {
		self.Pageno = "pageno"
	}
	if self.Totalno == "" {
		self.Totalno = "totalno"
	}
	if self.Rowcount == "" {
		self.Rowcount = "rowcount"
	}
	if self.Maxpageno == "" {
		self.Maxpageno = "maxpage"
	}
	if self.Fields == "" {
		self.Fields = "fields"
	}
}

// selectType returns variables' SELECT sql string and labels as []interface{}
// which can be used in SelectSQL
//
func selectType(selectPars map[string][2]string) (string, []interface{}) {
	if selectPars == nil {
		return "", nil
	}

	keys   := make([]string, 0)
	labels := make([]interface{}, 0)
	for key, val := range selectPars {
		keys = append(keys, key)
		labels = append(labels, val)
	}
	return strings.Join(keys, ", "), labels
}

// selectCondition returns the WHERE constraint
// 1) if key has single value, it means a simple EQUAL constraint
// 2) if key has array values, it mean an IN constrain
// 3) if key is "_gsql", it means a raw SQL statement.
// 4) it is the AND condition between keys.
//
func selectCondition(extra map[string]interface{}, table ...string) (string, []interface{}) {
	sql := ""
	values := make([]interface{}, 0)
	i := 0
	for field, valueInterface := range extra {
		if i > 0 {
			sql += " AND "
		}
		i++
		sql += "("

		if hasValue(table) {
			match, err := regexp.MatchString("\\.", field)
			if err == nil && !match {
				field = table[0] + "." + field
			}
		}
		switch value := valueInterface.(type) {
		case []int:
			n := len(value)
			sql += field + " IN (" + strings.Join(strings.Split(strings.Repeat("?", n), ""), ",") + ")"
			for _, v := range value {
				values = append(values, v)
			}
		case []int64:
			n := len(value)
			sql += field + " IN (" + strings.Join(strings.Split(strings.Repeat("?", n), ""), ",") + ")"
			for _, v := range value {
				values = append(values, v)
			}
		case []string:
			n := len(value)
			sql += field + " IN (" + strings.Join(strings.Split(strings.Repeat("?", n), ""), ",") + ")"
			for _, v := range value {
				values = append(values, v)
			}
		case string:
			if len(field) >= 5 && field[(len(field)-5):len(field)] == "_gsql" {
				sql += value
			} else {
				sql += field + " =?"
				values = append(values, value)
			}
		default:
			sql += field + " =?"
			values = append(values, value)
		}
		sql += ")"
	}

	return sql, values
}

func (self *Table) SetTopicsPars(newPars map[string][2]string) {
	self.topicsHashPars = newPars
}

func (self *Table) SetEditPars(newPars map[string][2]string) {
	self.editHashPars = newPars
}

// singleCondition returns WHERE constrains in existence of ids.
// If PK is a single column, ids should be a slice of targeted PK values
// To select a single PK equaling to 1234, just use ids = []int{1234}
// if PK has multiple columns, i.e. CurrentKeys exists, ids should be a slice of value arrays.
//
func (self *Table) singleCondition(ids []interface{}, extra ...map[string]interface{}) (string, []interface{}) {
	sql := ""
	extraValues := make([]interface{}, 0)

	if vs := self.CurrentKeys; hasValue(vs) {
		for i, item := range vs {
			val := ids[i]
			if i == 0 {
				sql = "("
			} else {
				sql += " AND "
			}
			switch idValues := val.(type) {
			case []interface{}:
				n := len(idValues)
				sql += item + " IN (" + strings.Join(strings.Split(strings.Repeat("?", n), ""), ",") + ")"
				for _, v := range idValues {
					extraValues = append(extraValues, v)
				}
			default:
				sql += item + " =?"
				extraValues = append(extraValues, val)
			}
		}
		sql += ")"
	}

	if hasValue(extra) && hasValue(extra[0]) {
		s, arr := selectCondition(extra[0])
		sql += " AND " + s
		for _, v := range arr {
			extraValues = append(extraValues, v)
		}
	}

	return sql, extraValues
}
