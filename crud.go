package godbi

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func (self *Model) insertHashContext(ctx context.Context, args map[string]interface{}) error {
	return self._insertHashContext(ctx, "INSERT", args)
}

func (self *Model) replaceHashContext(ctx context.Context, args map[string]interface{}) error {
	return self._insertHashContext(ctx, "REPLACE", args)
}

func (self *Model) _insertHashContext(ctx context.Context, how string, args map[string]interface{}) error {
	fields := make([]string, 0)
	values := make([]interface{}, 0)
	for k, v := range args {
		if v != nil {
			fields = append(fields, k)
			values = append(values, v)
		}
	}
	sql := how + " INTO " + self.CurrentTable + " (" + strings.Join(fields, ", ") + ") VALUES (" + strings.Join(strings.Split(strings.Repeat("?", len(fields)), ""), ",") + ")"
	return self.DoSQLContext(ctx, sql, values...)
}

func (self *Model) updateHashContext(ctx context.Context, args map[string]interface{}, ids []interface{}, extra ...map[string]interface{}) error {
	empties := make([]string, 0)
	return self.updateHashNullsContext(ctx, args, ids, empties, extra...)
}

func (self *Model) updateHashNullsContext(ctx context.Context, args map[string]interface{}, ids []interface{}, empties []string, extra ...map[string]interface{}) error {
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
		values = append(values, v)
	}

	sql := "UPDATE " + self.CurrentTable + " SET " + strings.Join(field0, ", ")
	for _, v := range empties {
		if args[v] != nil {
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

	return self.DoSQLContext(ctx, sql, values...)
}

func (self *Model) insupdTableContext(ctx context.Context, args map[string]interface{}, uniques []string) error {
	s := "SELECT " + self.CurrentKey + " FROM " + self.CurrentTable + "\nWHERE "
	v := make([]interface{}, 0)
	for i, val := range uniques {
		if i > 0 {
			s += " AND "
		}
		s += val + "=?"
		x := args[val]
		if x == nil {
			return errors.New("unique key value not found")
		}
		v = append(v, x)
	}

	lists := make([]map[string]interface{}, 0)
	if err := self.SelectContext(ctx, &lists, s, v...); err != nil {
		return err
	}
	if len(lists) > 1 {
		return errors.New("multiple records, not unique")
	}

	if len(lists) == 1 {
		id := lists[0][self.CurrentKey]
		if err := self.updateHashContext(ctx, args, []interface{}{id}); err != nil {
			return err
		}
		id64, err := strconv.ParseInt(fmt.Sprintf("%d", id), 10, 64)
		if err != nil {
			return err
		}
		self.LastID = id64
		self.Updated = true
	} else {
		if err := self.insertHashContext(ctx, args); err != nil {
			return err
		}
		self.Updated = false
	}

	return nil
}

func (self *Model) deleteHashContext(ctx context.Context, extra ...map[string]interface{}) error {
	sql := "DELETE FROM " + self.CurrentTable
	if !hasValue(extra) {
		return errors.New("delete whole table is not supported")
	}
	where, values := selectCondition(extra[0])
	if where != "" {
		sql += "\nWHERE " + where
	}
	return self.DoSQLContext(ctx, sql, values...)
}

func (self *Model) editHashContext(ctx context.Context, lists *[]map[string]interface{}, editPars map[string]interface{}, ids []interface{}, extra ...map[string]interface{}) error {
	sql, labels := selectType(editPars)
	sql = "SELECT " + sql + "\nFROM " + self.CurrentTable
	where, extraValues := self.singleCondition(ids, extra...)
	if where != "" {
		sql += "\nWHERE " + where
	}

	return self.SelectSQLContext(ctx, lists, labels, sql, extraValues...)
}

func (self *Model) topicsHashContext(ctx context.Context, lists *[]map[string]interface{}, selectPars map[string]interface{}, extra ...map[string]interface{}) error {
	return self.topicsHashOrderContext(ctx, lists, selectPars, "", extra...)
}

func (self *Model) topicsHashOrderContext(ctx context.Context, lists *[]map[string]interface{}, selectPars map[string]interface{}, order string, extra ...map[string]interface{}) error {
	sql, labels := selectType(selectPars)
	var table []string
	if hasValue(self.CurrentTables) {
		sql = "SELECT " + sql + "\nFROM " + joinString(self.CurrentTables)
		table = []string{self.CurrentTables[0].getAlias()}
	} else {
		sql = "SELECT " + sql + "\nFROM " + self.CurrentTable
	}

	if hasValue(extra) &&hasValue(extra[0]) {
		where, values := selectCondition(extra[0], table...)
		if where != "" {
			sql += "\nWHERE " + where
		}
		if order != "" {
			sql += "\n" + order
		}
		return self.SelectSQLContext(ctx, lists, labels, sql, values...)
	}

	if order != "" {
		sql += "\n" + order
	}
	return self.SelectSQLContext(ctx, lists, labels, sql)
}

func (self *Model) totalHashContext(ctx context.Context, v interface{}, extra ...map[string]interface{}) error {
	str := "SELECT COUNT(*) FROM " + self.CurrentTable

	if hasValue(extra) {
		where, values := selectCondition(extra[0])
		if where != "" {
			str += "\nWHERE " + where
		}
		return self.DB.QueryRowContext(ctx, str, values...).Scan(v)
	}

	return self.DB.QueryRowContext(ctx, str).Scan(v)
}
