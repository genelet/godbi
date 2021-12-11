package godbi

import (
	"context"
	"database/sql"
	"strings"
)

type Delecs struct {
	Action
}

func (self *Delecs) RunAction(db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
	return self.RunActionContext(context.Background(), db, t, ARGS, extra...)
}

func (self *Delecs) RunActionContext(ctx context.Context, db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
	labels := t.Pks
	if t.IdAuto != t.Pks[0] {
		labels = append(labels, t.IdAuto)
	}
	if t.Fks != nil && t.Fks[4] != t.IdAuto && t.Fks[4] != t.Pks[0] {
		labels = append(labels, t.Fks[4])
	}
	dbi := &DBI{DB: db}
	lists := make([]map[string]interface{}, 0)
	err := dbi.SelectContext(ctx, &lists, `SELECT ` + strings.Join(labels, ", ") + ` FROM ` + t.TableName + ` WHERE ` + t.Fks[3] + `=?`, ARGS[t.Fks[3]])
	return lists, err
}
