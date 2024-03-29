package godbi

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type DBType int64

const (
	SQLDefault DBType = iota
	SQLRaw
	SQLite
	MySQL
	Postgres
	TSMillisecond
	TSMicrosecond
)

type Col struct {
	ColumnName string  `json:"columnName" hcl:"columnName"`
	TypeName string    `json:"typeName" hcl:"typeName"`
	Label string       `json:"label" hcl:"label"`
	Notnull bool       `json:"notnull" hcl:"notnull"`
	Auto bool          `json:"auto" hcl:"auto"`
}

type Fk struct {
	FkTable  string    `json:"fkTable" hcl:"fkTable"`
	FkColumn string    `json:"fkColumn" hcl:"fkColumn"`
	Column   string    `json:"column" hcl:"column"`
}

type Table struct {
	TableName string   `json:"tableName" hcl:"tableName"`
    Columns   []*Col   `json:"columns" hcl:"columns"`
	Pks       []string `json:"pks,omitempty" hcl:"pks,optional"`
	IdAuto    string   `json:"idAuto,omitempty" hcl:"idAuto,optional"`
	Fks       []*Fk    `json:"fks,omitempty" hcl:"fks,optional"`
	Uniques   []string `json:"uniques,omitempty" hcl:"uniques,optional"`
	questionNumber DBType
}

func (self *Table) GetTableName() string {
	return self.TableName
}

func (self *Table) SetQuestionNumber(is DBType) {
	self.questionNumber = is
}

func (self *Table) RefreshArgs(args interface{}) interface{} {
	if args == nil { return args }

	cut := func(item map[string]interface{}) map[string]interface{} {
		for _, col := range self.Columns {
			if _, ok := item[col.ColumnName]; !ok {
				if v, ok := item[col.Label]; ok {
					item[col.ColumnName] = v
//					delete(item, col.Label)
				}
			}
		}
		return item
	}

	switch t := args.(type) {
	case []map[string]interface{}:
		var lists []map[string]interface{}
		for _, item := range t {
			lists = append(lists, cut(item))
		}
	case map[string]interface{}:
		return cut(t)
	default:
	}

	return nil
}

func (self *Table) getKeyColumns() []string {
	labels := make(map[string]bool)
	for _, pk := range self.Pks {
		labels[pk] = true
	}
	if self.IdAuto != "" {
		labels[self.IdAuto] = true
	}
	if self.Fks != nil {
		for _, fk := range self.Fks {
			labels[fk.Column] = true
		}
	}

	var outs []string
	for k := range labels {
		outs = append(outs, k)
	}
	return outs
}

func (self *Table) getFv(ARGS map[string]interface{}) map[string]interface{} {
    fieldValues := make(map[string]interface{})
    for f, l := range self.insertCols() {
        v, ok := ARGS[f]
		if !ok {
			v, ok = ARGS[l]
		}
		if ok {
			switch t := v.(type) {
			case []map[string]interface{}, map[string]interface{}:
			case bool:
				switch self.questionNumber {
				case SQLite, TSMillisecond, TSMicrosecond:
					if t {
						fieldValues[f] = 1
					} else {
						fieldValues[f] = 0
					}
				default:
					if t {
						fieldValues[f] = "true"
					} else {
						fieldValues[f] = "false"
					}
				}
			default:
				fieldValues[f] = t
			}
		}
    }
    return fieldValues
}

func (self *Table) checkNull(ARGS map[string]interface{}, extra ...map[string]interface{}) error {
	for _, col := range self.Columns {
		if col.Notnull == false || col.Auto == true {
			continue
		} // the column is ok with null
		err := fmt.Errorf("item %s not found in input", col.ColumnName)
		if _, ok := ARGS[col.ColumnName]; !ok {
			if hasValue(extra) && hasValue(extra[0]) {
				if _, ok = extra[0][col.ColumnName]; !ok {
					return err
				}
			} else {
				return err
			}
		}
	}
	return nil
}

func (self *Table) insertCols() map[string]string {
	cols := make(map[string]string)
	for _, col := range self.Columns {
		if col.Auto { continue }
		cols[col.ColumnName] = col.Label
	}
	return cols
}

func (self *Table) insertHashContext(ctx context.Context, db *sql.DB, args map[string]interface{}) (int64, error) {
	var fields []string
	var values []interface{}
	if self.IdAuto != "" && self.questionNumber == TSMillisecond {
		fields = append(fields, self.IdAuto)
		values = append(values, time.Now().UnixNano()/int64(time.Millisecond))
	} else if self.IdAuto != "" && self.questionNumber == TSMicrosecond {
		fields = append(fields, self.IdAuto)
		values = append(values, time.Now().UnixNano()/int64(time.Microsecond))
	}
	for k, v := range args {
		if v != nil {
			fields = append(fields, k)
			values = append(values, v)
		}
	}

	sql := "INSERT INTO " + self.TableName + " (" + strings.Join(fields, ", ") + ") VALUES (" + strings.Join(strings.Split(strings.Repeat("?", len(fields)), ""), ",") + ")"

	dbi := &DBI{DB: db}
	var err error
	switch self.questionNumber {
	case Postgres:
		sql = questionMarkerNumber(sql)
		if self.IdAuto != "" {
			sql += " RETURNING " + self.IdAuto
			err = dbi.InsertSerialContext(ctx, sql, values...)
		} else {
			err = dbi.DoSQLContext(ctx, sql, values...)
		}
	case SQLRaw, TSMillisecond, TSMicrosecond:
		err = dbi.DoSQLContext(ctx, sql, values...)
	default:
		err = dbi.InsertIDContext(ctx, sql, values...)
	}
	if err != nil {
		return 0, err
	}
	return dbi.LastID, nil
}

func (self *Table) updateHashNullsContext(ctx context.Context, db *sql.DB, args map[string]interface{}, ids []interface{}, empties []string, extra ...map[string]interface{}) error {
	if !hasValue(args) {
		return fmt.Errorf("no input data")
	}
	for _, k := range self.Pks {
		if grep(empties, k) {
			return fmt.Errorf("PK can't be NULL")
		}
	}

	var fields []string
	var field0 []string
	var values []interface{}
	for k, v := range args {
		fields = append(fields, k)
		field0 = append(field0, k+"=?")
		values = append(values, v)
	}

	sql := "UPDATE " + self.TableName + " SET " + strings.Join(field0, ", ")
	for _, v := range empties {
		if _, ok := args[v]; ok {
			continue
		}
		sql += ", " + v + "=NULL"
	}

	where, extraValues := self.singleCondition(ids, "", extra...)
	if where != "" {
		sql += "\nWHERE " + where
		for _, v := range extraValues {
			values = append(values, v)
		}
	}

	dbi := &DBI{DB: db}
	if self.questionNumber == Postgres { sql = questionMarkerNumber(sql) }
	return dbi.DoSQLContext(ctx, sql, values...)
}

func (self *Table) insupdTableContext(ctx context.Context, db *sql.DB, args map[string]interface{}) (int64, error) {
	changed := int64(0)
	s := "SELECT " + strings.Join(self.Pks, ", ") + " FROM " + self.TableName + "\nWHERE "
	var v []interface{}
	if self.Uniques == nil {
		return changed, fmt.Errorf("unique key not defined")
	}
	for i, val := range self.Uniques {
		if i > 0 {
			s += " AND "
		}
		s += val + "=?"
		if x, ok := args[val]; ok {
			v = append(v, x)
		} else {
			return changed, fmt.Errorf("input of unique key %s not found", val)
		}
	}

	lists := make([]map[string]interface{}, 0)
	dbi := &DBI{DB: db}
	if self.questionNumber == Postgres { s = questionMarkerNumber(s) }
	err := dbi.SelectContext(ctx, &lists, s, v...)
	if err != nil {
		return changed, err
	}
	if len(lists) > 1 {
		return changed, fmt.Errorf("multiple records found for unique key")
	}

	if len(lists) == 1 {
		ids := make([]interface{}, 0)
		for _, k := range self.Pks {
			ids = append(ids, lists[0][k])
		}
		err = self.updateHashNullsContext(ctx, db, args, ids, nil)
		if err == nil && self.IdAuto != "" {
			res := make(map[string]interface{})
			sql := "SELECT " + self.IdAuto + " FROM " + self.TableName + "\nWHERE " + strings.Join(self.Pks, "=? AND ") + "=?"
			if self.questionNumber == Postgres { sql = questionMarkerNumber(sql) }
			if err = dbi.GetSQLContext(ctx, res, sql, nil, ids...); err == nil {
				changed = res[self.IdAuto].(int64)
			}
		}
	} else {
		changed, err = self.insertHashContext(ctx, db, args)
	}

	return changed, err
}

func (self *Table) totalHashContext(ctx context.Context, db *sql.DB, v interface{}, extra ...map[string]interface{}) error {
	sql := "SELECT COUNT(*) FROM " + self.TableName

	if hasValue(extra) {
		where, values := selectCondition(extra[0], "")
		if where != "" {
			sql += "\nWHERE " + where
		}
		if self.questionNumber == Postgres { sql = questionMarkerNumber(sql) }
		return db.QueryRowContext(ctx, sql, values...).Scan(v)
	}

	return db.QueryRowContext(ctx, sql).Scan(v)
}

func (self *Table) getIdVal(ARGS map[string]interface{}, extra ...map[string]interface{}) []interface{} {
	if hasValue(extra) {
		return properValues(self.Pks, ARGS, extra[0])
	}
	return properValues(self.Pks, ARGS, nil)
}

func (self *Table) singleCondition(ids []interface{}, table string, extra ...map[string]interface{}) (string, []interface{}) {
	keys := self.Pks
	sql := ""
	var extraValues []interface{}

	for i, item := range keys {
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

	if hasValue(extra) && hasValue(extra[0]) {
		s, arr := selectCondition(extra[0], table)
		sql += " AND " + s
		for _, u := range arr {
			extraValues = append(extraValues, u)
		}
	}

	return sql, extraValues
}

func properValue(u string, ARGS map[string]interface{}, extra map[string]interface{}) interface{} {
	if !hasValue(extra) {
		return ARGS[u]
	}
	if val, ok := extra[u]; ok {
		return val
	}
	return ARGS[u]
}

func properValues(us []string, ARGS map[string]interface{}, extra map[string]interface{}) []interface{} {
	outs := make([]interface{}, len(us))
	if !hasValue(extra) {
		for i, u := range us {
			outs[i] = ARGS[u]
		}
		return outs
	}
	for i, u := range us {
		if val, ok := extra[u]; ok {
			outs[i] = val
		} else {
			outs[i] = ARGS[u]
		}
	}
	return outs
}

func properValuesHash(vs []string, ARGS map[string]interface{}, extra map[string]interface{}) map[string]interface{} {
	values := properValues(vs, ARGS, extra)
	hash := make(map[string]interface{})
	for i, v := range vs {
		hash[v] = values[i]
	}
	return hash
}

func selectCondition(extra map[string]interface{}, table string) (string, []interface{}) {
	sql := ""
	var values []interface{}
	i := 0

	for field, valueInterface := range extra {
		if i > 0 {
			sql += " AND "
		}
		i++
		sql += "("

		if table != "" {
			match, err := regexp.MatchString("\\.", field)
			if err == nil && !match {
				field = table + "." + field
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
			n := len(field)
			if n >= 5 && field[(n-5):] == "_gsql" {
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

func (self *Table) filterPars(ARGS map[string]interface{}, fieldsName string, joins []*Joint) (string, []interface{}, string) {
	var fields []string
	if v, ok := ARGS[fieldsName]; ok {
		fields = v.([]string)
	}

	var keys []string
	var labels []interface{}
	for _, col := range self.Columns {
		if fields==nil || grep (fields, col.ColumnName) {
			keys = append(keys, col.ColumnName)
			labels = append(labels, [2]string{col.Label, col.TypeName})
		}
	}
	sql := strings.Join(keys, ", ")

	var table string
	if hasValue(joins) {
		sql = "SELECT " + sql + "\nFROM " + joinString(joins)
		table = joins[0].getAlias()
	} else {
		sql = "SELECT " + sql + "\nFROM " + self.TableName
	}

	return sql, labels, table
}

func fromFv(fv map[string]interface{}) []map[string]interface{} {
	return []map[string]interface{}{fv}
}
