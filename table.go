package godbi

import (
	"encoding/json"
	"net/url"
	"regexp"
	"strings"
)

// Table describes a table used in multiple joined SELECT query.
//
type Join struct {
	// Name: the name of the table
	Name string `json:"name"`
	// Alias: the alias of the name
	Alias string `json:"alias,omitempty"`
	// Type: INNER or LEFT, how the table if joined
	Type string `json:"type,omitempty"`
	// Using: optional, join by USING column name.
	Using string `json:"using,omitempty"`
	// On: optional, join by ON condition
	On string `json:"on,omitempty"`
	// Sortby: optional, column to sort, only applied to the first table
	Sortby string `json:"sortby,omitempty"`
}

// joinString outputs the joined SQL statements from multiple tables.
//
func joinString(tables []*Join) string {
	sql := ""
	for i, table := range tables {
		name := table.Name
		if table.Alias != "" {
			name += " " + table.Alias
		}
		if i == 0 {
			sql = name
		} else if table.Using != "" {
			sql += "\n" + table.Type + " JOIN " + name + " USING (" + table.Using + ")"
		} else {
			sql += "\n" + table.Type + " JOIN " + name + " ON (" + table.On + ")"
		}
	}

	return sql
}

func (self *Join) getAlias() string {
	if self.Alias != "" {
		return self.Alias
	}
	return self.Name
}

// Page type describes next page's structure
// Model: the name of the model
// Action: the method name on the model
// Manual: constraint conditions manually assigned
// RelateItem: current page's column versus next page's column. The value is forced as constraint.
type Page struct {
	Model      string            `json:"model"`
	Action     string            `json:"action"`
	Manual     map[string]string `json:"manual,omitempty"`
	RelateItem map[string]string `json:"relate_item,omitempty"`
}

func (self *Page) refresh(item map[string]interface{}, extra url.Values) (url.Values, bool) {
	newExtra := extra
	found := false
	for k, v := range self.RelateItem {
		if t, ok := item[k]; ok {
			found = true
			newExtra.Set(v, interface2String(t))
			break
		}
	}
	return newExtra, found
}

// Table gives the RESTful table structure, usually parsed by JSON file on disk
//
type Table struct {
	// CurrentTable: the current table name
	CurrentTable string `json:"current_table,omitempty"`
	// CurrentTables: optional, for read-all SELECT with other joined tables
	CurrentTables []*Join `json:"current_tables,omitempty"`
	// CurrentKey: the single primary key of the table
	CurrentKey string `json:"current_key,omitempty"`
	// CurrentKeys: optional, if the primary key has multiple columns
	CurrentKeys []string `json:"current_keys,omitempty"`
	// CurrentIDAuto: if the table has an auto assigned series number
	CurrentIDAuto string `json:"current_id_auto,omitempty"`

	// Table columns for Crud
	// InsertPars: the columns used for Create
	InsertPars []string `json:"insert_pars,omitempty"`
	// EditPar: the columns used for Read One
	EditPars []string `json:"edit_pars,omitempty"`
	// UpdatePars: the columns used for Update
	UpdatePars []string `json:"update_pars,omitempty"`
	// InsupdPars: combination of the columns gives uniqueness
	InsupdPars []string `json:"insupd_pars,omitempty"`
	// TopicsPars: the columns used for Read All
	TopicsPars []string `json:"topics_pars,omitempty"`
	// TopicsHash: a map between SQL columns and output keys
	TopicsHash     map[string]json.RawMessage `json:"topics_hash,omitempty"`
	topicsHashPars map[string]string

	// TotalForce controls how the total number of rows be calculated for Topics
	// <-1	use ABS(TotalForce) as the total count
	// -1	always calculate the total count
	// 0	don't calculate the total count
	// 0	calculate only if the total count is not passed in args
	TotalForce int `json:"total_force,omitempty"`

	// Nextpages: defining how to call other models' actions
	Nextpages map[string][]*Page `json:"nextpages,omitempty"`

	Empties     string `json:"empties,omitempty"`
	Fields      string `json:"fields,omitempty"`
	Maxpageno   string `json:"maxpageno,omitempty"`
	Totalno     string `json:"totalno,omitempty"`
	Rowcount    string `json:"rawcount,omitempty"`
	Pageno      string `json:"pageno,omitempty"`
	Sortreverse string `json:"sortreverse,omitempty"`
	Sortby      string `json:"sortby,omitempty"`
}

// selectType returns variables' SELECT sql string, labels and types. 4 cases of interface{}
// []string{name}	just a list of column names
// [][2]string{name, type}	a list of column names and associated data types
// map[string]string{name: label}	rename the column names by labels
// map[string][2]string{name: label, type}	rename the column names to labels and use the specific types
//
func selectType(selectPars interface{}) (string, []string, []string) {
	switch vs := selectPars.(type) {
	case []string:
		labels := make([]string, 0)
		for _, v := range vs {
			labels = append(labels, v)
		}
		return strings.Join(labels, ", "), labels, nil
	case [][2]string:
		labels := make([]string, 0)
		types := make([]string, 0)
		for _, v := range vs {
			labels = append(labels, v[0])
			types = append(labels, v[1])
		}
		return strings.Join(labels, ", "), labels, types
	case map[string]string:
		labels := make([]string, 0)
		keys := make([]string, 0)
		for key, val := range vs {
			keys = append(keys, key)
			labels = append(labels, val)
		}
		return strings.Join(keys, ", "), labels, nil
	case map[string][2]string:
		labels := make([]string, 0)
		keys := make([]string, 0)
		types := make([]string, 0)
		for key, val := range vs {
			keys = append(keys, key)
			labels = append(labels, val[0])
			types = append(labels, val[1])
		}
		return strings.Join(keys, ", "), labels, types
	default:
	}
	return selectPars.(string), []string{selectPars.(string)}, nil
}

// selectCondition returns the WHERE constraint
// 1) if key has single value, it means a simple EQUAL constraint
// 2) if key has array values, it mean an IN constrain
// 3) if key is "_gsql", it means a raw SQL statement.
// 4) it is the AND condition between keys.
//
func selectCondition(extra url.Values, table ...string) (string, []interface{}) {
	sql := ""
	values := make([]interface{}, 0)
	i := 0
	for field, value := range extra {
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
		n := len(value)
		if n > 1 {
			sql += field + " IN (" + strings.Join(strings.Split(strings.Repeat("?", n), ""), ",") + ")"
			for _, v := range value {
				values = append(values, v)
			}
		} else {
			if len(field) >= 5 && field[(len(field)-5):len(field)] == "_gsql" {
				sql += value[0]
			} else {
				sql += field + " =?"
				values = append(values, value[0])
			}
		}
		sql += ")"
	}

	return sql, values
}

// singleCondition returns WHERE constrains in existence of ids.
// If PK is a single column, ids should be a slice of targeted PK values
// To select a single PK equaling to 1234, just use ids = []int{1234}
// if PK has multiple columns, i.e. CurrentKeys exists, ids should be a slice of value arrays.
//
func (self *Table) singleCondition(ids []interface{}, extra ...url.Values) (string, []interface{}) {
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
	} else {
		n := len(ids)
		if n > 1 {
			sql = "(" + self.CurrentKey + " IN (" + strings.Join(strings.Split(strings.Repeat("?", n), ""), ",") + "))"
		} else {
			sql = "(" + self.CurrentKey + "=?)"
		}
		for _, v := range ids {
			extraValues = append(extraValues, v)
		}
	}

	if hasValue(extra) {
		s, arr := selectCondition(extra[0])
		sql += " AND " + s
		for _, v := range arr {
			extraValues = append(extraValues, v)
		}
	}

	return sql, extraValues
}