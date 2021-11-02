package godbi

import (
	"context"
	"database/sql"
	"fmt"
)

type Delete struct {
	Action
}

func (self *Delete) RunAction(db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, []*Edge, error) {
    return self.RunActionContext(context.Background(), db, t, ARGS, extra...)
}

func (self *Delete) RunActionContext(ctx context.Context, db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, []*Edge, error) {
	err := self.checkNull(ARGS)
	if err != nil { return nil, nil, err }

	ids := t.getIdVal(ARGS, extra...)
	if !hasValue(ids) {
		return nil, nil, fmt.Errorf("pk value not provided")
	}

    sql := "DELETE FROM " + t.CurrentTable
	where, values := t.singleCondition(ids, "", extra...)
	if where != "" {
        sql += "\nWHERE " + where
    } else {
        return nil, nil, fmt.Errorf("delete whole table is not supported")
    }
	dbi := &DBI{DB:db}
    return extra, self.Nextpages, dbi.DoSQLContext(ctx, sql, values...)
}
