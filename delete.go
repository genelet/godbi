package godbi

import (
	"context"
	"database/sql"
	"fmt"
)

type Delete struct {
	Action
}

func (self *Delete) RunAction(db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]interface{}, error) {
	return self.RunActionContext(context.Background(), db, t, ARGS, extra...)
}

func (self *Delete) RunActionContext(ctx context.Context, db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]interface{}, error) {
	ids := t.getIdVal(ARGS, extra...)
	if !hasValue(ids) {
		return nil, fmt.Errorf("pk value not provided")
	}

	sql := "DELETE FROM " + t.TableName
	where, values := t.singleCondition(ids, "", extra...)
	if where != "" {
		sql += "\nWHERE " + where
	} else {
		return nil, fmt.Errorf("delete whole table is not supported")
	}
	dbi := &DBI{DB: db}
	return nil, dbi.DoSQLContext(ctx, sql, values...)
}
