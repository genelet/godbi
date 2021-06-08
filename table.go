package godbi

import (
	"encoding/json"
	"regexp"
	"strings"
)

// Table describes a table used in multiple joined SELECT query.
//
type Join struct {
	// Name: the name of the table
	Name string `json:"name" hcl:"name,label"`
	// Alias: the alias of the name
	Alias string `json:"alias,omitempty" hcl:"alias,optional"`
	// Type: INNER or LEFT, how the table if joined
	Type string `json:"type,omitempty" hcl:"type,optional"`
	// Using: optional, join by USING column name.
	Using string `json:"using,omitempty" hcl:"using,optional"`
	// On: optional, join by ON condition
	On string `json:"on,omitempty" hcl:"on,optional"`
	// Sortby: optional, column to sort, only applied to the first table
	Sortby string `json:"sortby,omitempty" hcl:"sortby,optional"`
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
	Model      string            `json:"model" hcl:"model,label"`
	Action     string            `json:"action" hcl:"action,label"`
	Extra      map[string]string `json:"extra,omitempty" hcl:"extra,optional"`
	RelateItem map[string]string `json:"relate_item,omitempty" hcl:"relate_item"`
}

func (self *Page) refresh(item, extra map[string]interface{}) (map[string]interface{}, bool) {
	newExtra := make(map[string]interface{})
	for k, v := range extra {
		newExtra[k] = v
	}
	found := false
	for k, v := range self.RelateItem {
		if t, ok := item[k]; ok {
			found = true
			newExtra[v] = t
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
	EditPars []interface{} `json:"edit_pars,omitempty"`
	// EditHash: the columns used for Read One
	EditHash map[string]interface{} `json:"edit_hash,omitempty"`
	// UpdatePars: the columns used for Update
	UpdatePars []string `json:"update_pars,omitempty"`
	Empties    []string `json:"empties,omitempty"`
	// InsupdPars: combination of the columns gives uniqueness
	InsupdPars []string `json:"insupd_pars,omitempty"`
	// TopicsPars: the columns used for Read All
	TopicsPars []interface{} `json:"topics_pars,omitempty"`
	// TopicsHash: a map between SQL columns and output keys
	TopicsHash map[string]interface{} `json:"topics_hash,omitempty"`

	editHashPars interface{}
	topicsHashPars interface{}

	// TotalForce controls how the total number of rows be calculated for Topics
	// <-1	use ABS(TotalForce) as the total count
	// -1	always calculate the total count
	// 0	don't calculate the total count
	// 1	calculate only if the total count is not passed in args
	TotalForce int `json:"total_force,omitempty"`

	// Nextpages: defining how to call other models' actions
	Nextpages map[string][]*Page `json:"nextpages,omitempty"`

	Fields      string `json:"fields,omitempty"`
	Maxpageno   string `json:"maxpageno,omitempty"`
	Totalno     string `json:"totalno,omitempty"`
	Rowcount    string `json:"rawcount,omitempty"`
	Pageno      string `json:"pageno,omitempty"`
	Sortreverse string `json:"sortreverse,omitempty"`
	Sortby      string `json:"sortby,omitempty"`
}

func generalHashPars(TopicsHash map[string]interface{}, TopicsPars []interface{}, fields []string) interface{} {
	if hasValue(TopicsHash) {
		s2 := make(map[string][2]string)
		s1 := make(map[string]string)
		for k, vs := range TopicsHash {
			if fields != nil && !grep(fields, k) {
				continue
			}
			switch v := vs.(type) {
			case []interface{}:
				s2[k] = [2]string{v[0].(string), v[1].(string)}
			default:
				s1[k] = v.(string)
			}
		}
		if len(s2) > 0 {
			return s2
		}
		return s1
	}

	s2 := make([][2]string,0)
	s1 := make([]string,0)
	for _, vs := range TopicsPars {
		switch v := vs.(type) {
		case []interface{}:
			if fields != nil && !grep(fields, v[0].(string)) {
				continue
			}
			s2 = append(s2, [2]string{v[0].(string), v[1].(string)})
		default:
			if fields != nil && !grep(fields, v.(string)) {
				continue
			}
			s1 = append(s1, v.(string))
		}
	}
	if len(s2) > 0 {
		return s2
	}
	return s1
}

func filterPars(selectPars interface{}, fields []string) interface{} {
	if fields == nil || selectPars == nil {
		return nil
	}

	switch vs := selectPars.(type) {
	case []string:
		labels := make([]string, 0)
		for _, v := range vs {
			if grep(fields, v) {
				labels = append(labels, v)
			}
		}
		return labels
	case [][2]string:
		labels := make([][2]string, 0)
		for _, v := range vs {
			if grep(fields, v[0]) {
				labels = append(labels, v)
			}
		}
		return labels
	case map[string]string:
		labels := make(map[string]string, 0)
		for key, val := range vs {
			if grep(fields, val) {
				labels[key] = val
			}
		}
		return labels
	case map[string][2]string:
		labels := make(map[string][2]string, 0)
		for key, val := range vs {
			if grep(fields, val[0]) {
				labels[key] = val
			}
		}
		return labels
	default:
	}
	return nil
}

func newTable(content []byte) (*Table, error) {
	var parsed *Table
	if err := json.Unmarshal(content, &parsed); err != nil {
		return nil, err
	}

	parsed.topicsHashPars = generalHashPars(parsed.TopicsHash, parsed.TopicsPars, nil)
	parsed.editHashPars   = generalHashPars(parsed.EditHash, parsed.EditPars, nil)

	if parsed.Sortby == "" {
		parsed.Sortby = "sortby"
	}
	if parsed.Sortreverse == "" {
		parsed.Sortreverse = "sortreverse"
	}
	if parsed.Pageno == "" {
		parsed.Pageno = "pageno"
	}
	if parsed.Totalno == "" {
		parsed.Totalno = "totalno"
	}
	if parsed.Rowcount == "" {
		parsed.Rowcount = "rowcount"
	}
	if parsed.Maxpageno == "" {
		parsed.Maxpageno = "maxpage"
	}
	if parsed.Fields == "" {
		parsed.Fields = "fields"
	}

	return parsed, nil
}

// selectType returns variables' SELECT sql string, labels and types. 4 cases of interface{}
// []string{name}	just a list of column names
// [][2]string{name, type}	a list of column names and associated data types
// map[string]string{name: label}	rename the column names by labels
// map[string][2]string{name: label, type}	rename the column names to labels and use the specific types
//
func selectType(selectPars interface{}) (string, []string, []string) {
	if selectPars == nil {
		return "", nil, nil
	}

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

func (self *Table) SetTopicsPars(newPars interface{}) {
	self.topicsHashPars = newPars
}

func (self *Table) SetEditPars(newPars interface{}) {
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

	if hasValue(extra) && hasValue(extra[0]) {
		s, arr := selectCondition(extra[0])
		sql += " AND " + s
		for _, v := range arr {
			extraValues = append(extraValues, v)
		}
	}

	return sql, extraValues
}
