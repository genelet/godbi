package godbi

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// insertHash inserts one row into the table.
// args: the input row data expressed as url.Values.
// The keys are column names, and their values are columns' values.
//
func (self *Model) insertHash(args url.Values) error {
	return self._insertHash("INSERT", args)
}

// replaceHash inserts one row as using 'REPLACE' instead of 'INSERT'
// args: the input row data expressed as url.Values.
// The keys are column names, and their values are columns' values.
//
func (self *Model) replaceHash(args url.Values) error {
	return self._insertHash("REPLACE", args)
}

func (self *Model) _insertHash(how string, args url.Values) error {
	fields := make([]string, 0)
	values := make([]interface{}, 0)
	for k, v := range args {
		if v != nil {
			fields = append(fields, k)
			values = append(values, v[0])
		}
	}
	sql := how + " INTO " + self.CurrentTable + " (" + strings.Join(fields, ", ") + ") VALUES (" + strings.Join(strings.Split(strings.Repeat("?", len(fields)), ""), ",") + ")"
	return self.DoSQL(sql, values...)
}

// updateHash updates multiple rows using data expressed in type Values.
// args: columns names and their new values.
// ids: primary key's value, either a single value or array of values.
// extra: optional, extra constraints put on row's WHERE statement.
//
func (self *Model) updateHash(args url.Values, ids []interface{}, extra ...url.Values) error {
	empties := make([]string, 0)
	return self.updateHashNulls(args, ids, empties, extra...)
}

// updateHashNull updates multiple rows using data expressed in type Values.
// args: columns names and their new values.
// empties: if these columns have no values in args, they are forced to be NULL.
// ids: primary key's value, either a single value or array of values.
// extra: optional, extra constraints on WHERE statement.
//
func (self *Model) updateHashNulls(args url.Values, ids []interface{}, empties []string, extra ...url.Values) error {
	if empties == nil {
		empties = make([]string, 0)
	}
	if hasValue(self.CurrentKeys) {
		for _, k := range self.CurrentKeys {
			if grep(empties, k) {
				return errors.New("PK can't be NULL")
			}
		}
	} else if grep(empties, self.CurrentKey) {
		return errors.New("PK can't be NULL")
	}

	fields := make([]string, 0)
	field0 := make([]string, 0)
	values := make([]interface{}, 0)
	for k, v := range args {
		fields = append(fields, k)
		field0 = append(field0, k+"=?")
		values = append(values, v[0])
	}

	sql := "UPDATE " + self.CurrentTable + " SET " + strings.Join(field0, ", ")
	for _, v := range empties {
		if args.Get(v) != "" {
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

// insupdTable update a row if it is found to exists, or to inserts a new row
// args: row's column names and values
// uniques: combination of one or multiple columns to assert uniqueness.
//
func (self *Model) insupdTable(args url.Values, uniques []string) error {
	s := "SELECT " + self.CurrentKey + " FROM " + self.CurrentTable + "\nWHERE "
	v := make([]interface{}, 0)
	for i, val := range uniques {
		if i > 0 {
			s += " AND "
		}
		s += val + "=?"
		x := args.Get(val)
		if x == "" {
			return errors.New("unique key value not found")
		}
		v = append(v, x)
	}

	lists := make([]map[string]interface{}, 0)
	if err := self.SelectSQL(&lists, s, v...); err != nil {
		return err
	}
	if len(lists) > 1 {
		return errors.New("multiple records, not unique")
	}

	if len(lists) == 1 {
		id := lists[0][self.CurrentKey]
		if err := self.updateHash(args, []interface{}{id}); err != nil {
			return err
		}
		id64, err := strconv.ParseInt(fmt.Sprintf("%d", id), 10, 64)
		if err != nil {
			return err
		}
		self.LastID = id64
		self.Updated = true
	} else {
		if err := self.insertHash(args); err != nil {
			return err
		}
		self.Updated = false
	}

	return nil
}

// deleteHash deletes rows by extra: constraints on WHERE statement.
//
func (self *Model) deleteHash(extra ...url.Values) error {
	sql := "DELETE FROM " + self.CurrentTable
	if !hasValue(extra) {
		return errors.New("delete whole table is not supported")
	}
	where, values := selectCondition(extra[0])
	if where != "" {
		sql += "\nWHERE " + where
	}
	return self.DoSQL(sql, values...)
}

// editHash selects one or multiple rows from the primary key.
// lists: received the query results in slice of maps.
// ids: primary key values array.
// extra: optional, extra constraints on WHERE statement.
// editPars: defining which and how columns are returned:
// 1) []string{name} - name of column
// 2) [2]string{name, type} - name and data type of column
// 3) map[string]string{name: label} - column name is mapped to label
// 4) map[string][2]string{name: label, type} -- column name to label and data type
//
func (self *Model) editHash(lists *[]map[string]interface{}, editPars interface{}, ids []interface{}, extra ...url.Values) error {
	sql, labels, types := selectType(editPars)
	sql = "SELECT " + sql + "\nFROM " + self.CurrentTable
	where, extraValues := self.singleCondition(ids, extra...)
	if where != "" {
		sql += "\nWHERE " + where
	}

	return self.SelectSQLTypeLabel(lists, types, labels, sql, extraValues...)
}

// topicsHash selects all rows.
// lists: received the query results in slice of maps.
// extra: optional, extra constraints on WHERE statement.
// topicsPars: defining which and how columns are returned:
// 1) []string{name} - name of column
// 2) [2]string{name, type} - name and data type of column
// 3) map[string]string{name: label} - column name is mapped to label
// 4) map[string][2]string{name: label, type} -- column name to label and data type
//
func (self *Model) topicsHash(lists *[]map[string]interface{}, selectPars interface{}, extra ...url.Values) error {
	return self.topicsHashOrder(lists, selectPars, "", extra...)
}

// topicsHashOrder is the same as topicsHash with the order string
// order: a string like 'ORDER BY ...'
//
func (self *Model) topicsHashOrder(lists *[]map[string]interface{}, selectPars interface{}, order string, extra ...url.Values) error {
	sql, labels, types := selectType(selectPars)
	var table []string
	if hasValue(self.CurrentTables) {
		sql = "SELECT " + sql + "\nFROM " + joinString(self.CurrentTables)
		table = []string{self.CurrentTables[0].getAlias()}
	} else {
		sql = "SELECT " + sql + "\nFROM " + self.CurrentTable
	}

	if hasValue(extra) {
		where, values := selectCondition(extra[0], table...)
		if where != "" {
			sql += "\nWHERE " + where
		}
		if order != "" {
			sql += "\n" + order
		}
		return self.SelectSQLTypeLabel(lists, types, labels, sql, values...)
	}

	if order != "" {
		sql += "\n" + order
	}
	return self.SelectSQLTypeLabel(lists, types, labels, sql)
}

// totalHash returns the total number of rows available
// This function is used for pagination.
// v: the total number is returned in this referenced variable
// extra: optional, extra constraints on WHERE statement.
//
func (self *Model) totalHash(v interface{}, extra ...url.Values) error {
	str := "SELECT COUNT(*) FROM " + self.CurrentTable

	if hasValue(extra) {
		where, values := selectCondition(extra[0])
		if where != "" {
			str += "\nWHERE " + where
		}
		return self.DB.QueryRow(str, values...).Scan(v)
	}

	return self.DB.QueryRow(str).Scan(v)
}

/*
// Crud works on one table's CRUD or RESTful operations:
// C: create a new row
// R: read all rows, or read one row
// U: update a row
// D: delete a row
//
type Crud struct {
	DBI
	// CurrentTable: the current table name
	CurrentTable string `json:"current_table,omitempty"`
	// CurrentTables: optional, for read-all SELECT with other joined tables
	CurrentTables []*Join `json:"current_tables,omitempty"`
	// CurrentKey: the single primary key of the table
	CurrentKey string `json:"current_key,omitempty"`
	// CurrentKeys: optional, if the primary key has multiple columns
	CurrentKeys []string `json:"current_keys,omitempty"`
	// Updated: for Insupd only, indicating if the row is updated or new
	Updated bool
}

// newCrud creates a new Crud struct.
// db: the DB handle
// table: the name of table
// currentKey: the column name of the primary key
// currentKeys: if the PK consists of multiple columns, as []string
// tables: array of joined tables for real-all SELECT
//
func newCrud(db *sql.DB, table, currentKey string, tables []*Join, currentKeys []string) *Crud {
	crud := new(Crud)
	crud.DB = db
	crud.CurrentTable = table
	if tables != nil {
		crud.CurrentTables = tables
	}
	crud.CurrentKey = currentKey
	if currentKeys != nil {
		crud.CurrentKeys = currentKeys
	}
	return crud
}
*/

func newCrud(db *sql.DB, table, currentKey string, tables []*Join, currentKeys []string) *Model {
	crud := new(Model)
	crud.DB = db
	crud.CurrentTable = table
	if tables != nil {
		crud.CurrentTables = tables
	}
	crud.CurrentKey = currentKey
	if currentKeys != nil {
		crud.CurrentKeys = currentKeys
	}
	return crud
}
