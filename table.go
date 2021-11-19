package godbi

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

type Table struct {
	TableName string   `json:"table" hcl:"table"`
	Pks       []string `json:"pks,omitempty" hcl:"pks,optional"`
	IDAuto    string   `json:"id_auto,omitempty" hcl:"id_auto,optional"`
	Fks       []string `json:"fks,omitempty" hcl:"fks,optional"`
}

func (self *Table) insertHashContext(ctx context.Context, db *sql.DB, args map[string]interface{}) (int64, error) {
	fields := make([]string, 0)
	values := make([]interface{}, 0)
	for k, v := range args {
		if v != nil {
			fields = append(fields, k)
			values = append(values, v)
		}
	}
	sql := "INSERT INTO " + self.TableName + " (" + strings.Join(fields, ", ") + ") VALUES (" + strings.Join(strings.Split(strings.Repeat("?", len(fields)), ""), ",") + ")"
	dbi := &DBI{DB: db}
	err := dbi.DoSQLContext(ctx, sql, values...)
	if err != nil {
		return 0, err
	}
	return dbi.LastID, nil
}

func (self *Table) updateHashNullsContext(ctx context.Context, db *sql.DB, args map[string]interface{}, ids []interface{}, empties []string, extra ...map[string]interface{}) error {
	if empties == nil {
		empties = make([]string, 0)
	}
	for _, k := range self.Pks {
		if grep(empties, k) {
			return fmt.Errorf("PK can't be NULL")
		}
	}

	fields := make([]string, 0)
	field0 := make([]string, 0)
	values := make([]interface{}, 0)
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
	return dbi.DoSQLContext(ctx, sql, values...)
}

func (self *Table) insupdTableContext(ctx context.Context, db *sql.DB, args map[string]interface{}, uniques []string) (int64, error) {
	changed := int64(0)
	s := "SELECT " + strings.Join(self.Pks, ", ") + " FROM " + self.TableName + "\nWHERE "
	v := make([]interface{}, 0)
	for i, val := range uniques {
		if i > 0 {
			s += " AND "
		}
		s += val + "=?"
		if x, ok := args[val]; ok {
			v = append(v, x)
		} else {
			return changed, fmt.Errorf("unique key %s not found in input", val)
		}
	}

	lists := make([]map[string]interface{}, 0)
	dbi := &DBI{DB: db}
	err := dbi.SelectContext(ctx, &lists, s, v...)
	if err != nil {
		return changed, err
	}
	if len(lists) > 1 {
		return changed, fmt.Errorf("multiple records found")
	}

	if len(lists) == 1 {
		ids := make([]interface{}, 0)
		for _, k := range self.Pks {
			ids = append(ids, lists[0][k])
		}
		err = self.updateHashNullsContext(ctx, db, args, ids, nil)
	} else {
		changed, err = self.insertHashContext(ctx, db, args)
	}

	return changed, err
}

func (self *Table) totalHashContext(ctx context.Context, db *sql.DB, v interface{}, extra ...map[string]interface{}) error {
	str := "SELECT COUNT(*) FROM " + self.TableName

	if hasValue(extra) {
		where, values := selectCondition(extra[0], "")
		if where != "" {
			str += "\nWHERE " + where
		}
		return db.QueryRowContext(ctx, str, values...).Scan(v)
	}

	return db.QueryRowContext(ctx, str).Scan(v)
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
	extraValues := make([]interface{}, 0)

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
		for _, v := range arr {
			extraValues = append(extraValues, v)
		}
	}

	return sql, extraValues
}

func properValue(v string, ARGS, extra map[string]interface{}) interface{} {
	if !hasValue(extra) {
		return ARGS[v]
	}
	if val, ok := extra[v]; ok {
		return val
	}
	return ARGS[v]
}

func properValues(vs []string, ARGS, extra map[string]interface{}) []interface{} {
	outs := make([]interface{}, len(vs))
	if !hasValue(extra) {
		for i, v := range vs {
			outs[i] = ARGS[v]
		}
		return outs
	}
	for i, v := range vs {
		if val, ok := extra[v]; ok {
			outs[i] = val
		} else {
			outs[i] = ARGS[v]
		}
	}
	return outs
}

func properValuesHash(vs []string, ARGS, extra map[string]interface{}) map[string]interface{} {
	values := properValues(vs, ARGS, extra)
	hash := make(map[string]interface{})
	for i, v := range vs {
		hash[v] = values[i]
	}
	return hash
}

func selectCondition(extra map[string]interface{}, table string) (string, []interface{}) {
	sql := ""
	values := make([]interface{}, 0)
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
