package godbi

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Delecs struct {
	Action
}

func (self *Delecs) RunAction(db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]interface{}, error) {
	return self.RunActionContext(context.Background(), db, t, ARGS, extra...)
}

func (self *Delecs) RunActionContext(ctx context.Context, db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]interface{}, error) {
	dbi := &DBI{DB: db}
	lists := make([]interface{}, 0)
	if t.Fks == nil {
		return nil, fmt.Errorf("fks not define in %s", t.TableName)
	}
	str := ""
	var values []interface{}
	for _, fk := range t.Fks {
		name := fk.Column
		if v, ok := ARGS[name]; ok {
			if str != "" { str += ", " }
			str += name + "=?"
			values = append(values, v)
		}
	}
	if values == nil {
		return nil, fmt.Errorf("fks valeus not found in %s", t.TableName)
	}
	err := dbi.SelectContext(ctx, &lists, `SELECT ` + strings.Join(t.getKeyColumns(), ", ") + ` FROM ` + t.TableName + ` WHERE ` + str, values...)
	return lists, err
}
