package godbi

import (
	"database/sql"
	"errors"
	"net/url"
	"regexp"
	"strings"
)

// Crud works on table's REST actions
// CurrentTable: the current table name
// CurrentKey: the single primary key of the table
// CurrentKeys: if the primary key has multiple columns
// CurrentRow: the row corresponding to LastID
// Updated: true if the CurrentRow is old, false if it is new
type Crud struct {
	DBI
	CurrentTable  string   `json:"current_table,omitempty"`
	CurrentTables []*Table `json:"current_tables,omitempty"`
	CurrentKey    string   `json:"current_key,omitempty"`
	CurrentKeys   []string `json:"current_keys,omitempty"`
	Updated       bool
}

// NewCrud creates a new Crud struct.
// db: the DB handle
// table: the name of table
// key: the column name of the primary key
// tables: array of joint tables as the Table struct
// keys: if the PK consists of multiple columns, assign them as []string
func NewCrud(db *sql.DB, table, key string, tables []*Table, keys []string) *Crud {
	crud := new(Crud)
	crud.Db = db
	crud.CurrentTable = table
	if tables != nil {
		crud.CurrentTables = tables
	}
	crud.CurrentKey = key
	if keys != nil {
		crud.CurrentKeys = keys
	}
	return crud
}

func selectType(selectPars interface{}) (string, []string, []string) {
	switch vs := selectPars.(type) {
	case []string:
		labels := make([]string, 0)
		for _, v := range vs {
			labels = append(labels, v)
		}
		return strings.Join(labels, ", "), labels, nil
	case map[string]string:
		labels := make([]string, 0)
		types := make([]string, 0)
		for key, val := range vs {
			labels = append(labels, key)
			types = append(types, val)
		}
		return strings.Join(labels, ", "), labels, types
	default:
	}
	return selectPars.(string), []string{selectPars.(string)}, nil
}

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

		if HasValue(table) {
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

func (self *Crud) singleCondition(ids []interface{}, extra ...url.Values) (string, []interface{}) {
	sql := ""
	extraValues := make([]interface{}, 0)

	if vs := self.CurrentKeys; HasValue(vs) {
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

	if HasValue(extra) {
		s, arr := selectCondition(extra[0])
		sql += " AND " + s
		for _, v := range arr {
			extraValues = append(extraValues, v)
		}
	}

	return sql, extraValues
}

// InsertHash inserts a row as map (hash) into the table.
func (self *Crud) InsertHash(fieldValues url.Values) error {
	return self.insertHash("INSERT", fieldValues)
}

// ReplaceHash inserts a row as map (hash) into the table using 'REPLACE'
func (self *Crud) ReplaceHash(fieldValues url.Values) error {
	return self.insertHash("REPLACE", fieldValues)
}

func (self *Crud) insertHash(how string, fieldValues url.Values) error {
	fields := make([]string, 0)
	values := make([]interface{}, 0)
	for k, v := range fieldValues {
		if v != nil {
			fields = append(fields, k)
			values = append(values, v[0])
		}
	}
	sql := how + " INTO " + self.CurrentTable + " (" + strings.Join(fields, ", ") + ") VALUES (" + strings.Join(strings.Split(strings.Repeat("?", len(fields)), ""), ",") + ")"
	return self.DoSQL(sql, values...)
}

// UpdateHash updates a row as map (hash) into the table.
func (self *Crud) UpdateHash(fieldValues url.Values, ids []interface{}, extra ...url.Values) error {
	empties := make([]string, 0)
	return self.UpdateHashNulls(fieldValues, ids, empties, extra...)
}

// UpdateHashNulls updates a row as map (hash) into the table.
// empties: these columns will be forced to have default NULL.
func (self *Crud) UpdateHashNulls(fieldValues url.Values, ids []interface{}, empties []string, extra ...url.Values) error {
	if empties == nil {
		empties = make([]string, 0)
	}
	if HasValue(self.CurrentKeys) {
		for _, k := range self.CurrentKeys {
			if Grep(empties, k) {
				return errors.New("Assgin empties to key")
			}
		}
	} else if Grep(empties, self.CurrentKey) {
		return errors.New("Assgin empties to key")
	}

	fields := make([]string, 0)
	field0 := make([]string, 0)
	values := make([]interface{}, 0)
	for k, v := range fieldValues {
		fields = append(fields, k)
		field0 = append(field0, k+"=?")
		values = append(values, v[0])
	}

	sql := "UPDATE " + self.CurrentTable + " SET " + strings.Join(field0, ", ")
	for _, v := range empties {
		if fieldValues.Get(v) != "" {
			continue
		}
		sql += ", " + v + "=NULL"
	}

	where, extraValues := self.singleCondition(ids, extra...)
	if where != "" {
		sql += "\nWHERE " + where
	}
	for _, v := range extraValues {
		values = append(values, v)
	}

	return self.DoSQL(sql, values...)
}

// InsupdTable inserts a new row, or retrieves the old one depending on
// wether or not the row of the unique combinated columns 'uniques', exists.
func (self *Crud) InsupdTable(fieldValues url.Values, uniques []string) error {
	s := "SELECT " + self.CurrentKey + " FROM " + self.CurrentTable + "\nWHERE "
	v := make([]interface{}, 0)
	for i, val := range uniques {
		if i > 0 {
			s += " AND "
		}
		s += val + "=?"
		x := fieldValues.Get(val)
		if x == "" {
			return errors.New("1075")
		}
		v = append(v, x)
	}

	lists := make([]map[string]interface{}, 0)
	if err := self.SelectSQL(&lists, s, v...); err != nil {
		return err
	}
	if len(lists) > 1 {
		return errors.New("1070")
	}

	if len(lists) == 1 {
		id := lists[0][self.CurrentKey]
		if err := self.UpdateHash(fieldValues, []interface{}{id}); err != nil {
			return err
		}
		self.Updated = true
	} else {
		if err := self.InsertHash(fieldValues); err != nil {
			return err
		}
		self.Updated = false
	}

	return nil
}

// DeleteHash deletes having key in ids.
// The constrain is described in extra. See 'extra' in TopicsHash.
func (self *Crud) DeleteHash(ids []interface{}, extra ...url.Values) error {
	sql := "DELETE FROM " + self.CurrentTable
	where, extraValues := self.singleCondition(ids, extra...)
	if where != "" {
		sql += "\nWHERE " + where
	}
	return self.DoSQL(sql, extraValues...)
}

// EditHash selects rows having key in ids
// Only will columns defined in editPars will be returned.
// The select restriction is described in extra. See 'extra' in TopicsHash.
func (self *Crud) EditHash(lists *[]map[string]interface{}, editPars interface{}, ids []interface{}, extra ...url.Values) error {
	sql, labels, types := selectType(editPars)
	sql = "SELECT " + sql + "\nFROM " + self.CurrentTable
	where, extraValues := self.singleCondition(ids, extra...)
	if where != "" {
		sql += "\nWHERE " + where
	}

	return self.SelectSQLTypeLabel(lists, types, labels, sql, extraValues...)
}

// TopicsHash selects rows using restriction 'extra'.
// Only will columns defined in selectPars will be returned.
// Currently only will the following three tpyes os restrictions are support:
// key=>value: the key has the value
// key=>slice: the key has one of the values in the slice
// '_gsql'=>'raw sql statement': use the raw SQL statement
func (self *Crud) TopicsHash(lists *[]map[string]interface{}, selectPars interface{}, extra ...url.Values) error {
	return self.TopicsHashOrder(lists, selectPars, "", extra...)
}

// TopicsHashOrder is the same as TopicsHash, but use the order string as 'ORDER BY order'
func (self *Crud) TopicsHashOrder(lists *[]map[string]interface{}, selectPars interface{}, order string, extra ...url.Values) error {
	sql, labels, types := selectType(selectPars)
	var table []string
	if HasValue(self.CurrentTables) {
		sql = "SELECT " + sql + "\nFROM " + TableString(self.CurrentTables)
		table = []string{self.CurrentTables[0].getAlias()}
	} else {
		sql = "SELECT " + sql + "\nFROM " + self.CurrentTable
	}

	if HasValue(extra) {
		where, values := selectCondition(extra[0], table...)
		if where != "" {
			sql += "\nWHERE " + where
		}
		if order != "" {
			sql += "\nORDER BY " + order
		}
		return self.SelectSQLTypeLabel(lists, types, labels, sql, values...)
	}

	if order != "" {
		sql += "\nORDER BY " + order
	}
	return self.SelectSQLTypeLabel(lists, types, labels, sql)
}

// TotalHash returns the total number of rows available
// This function is used for pagination.
func (self *Crud) TotalHash(v interface{}, extra ...url.Values) error {
	str := "SELECT COUNT(*) FROM " + self.CurrentTable

	if HasValue(extra) {
		where, values := selectCondition(extra[0])
		if where != "" {
			str += "\nWHERE " + where
		}
		return self.Db.QueryRow(str, values...).Scan(v)
	}

	return self.Db.QueryRow(str).Scan(v)
}
