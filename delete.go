package godbi

import (
	"context"
	"database/sql"
	"fmt"
)

type Delete struct {
	Action
}

func (self *Delete) RunAction(db *sql.DB, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, []*Page, error) {
    return self.RunActionContext(context.Background(), db, ARGS, extra...)
}

func (self *Delete) RunActionContext(ctx context.Context, db *sql.DB, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, []*Page, error) {
	err := self.checkNull(ARGS)
	if err != nil { return nil, nil, err }

	ids := self.getIdVal(ARGS, extra...)
	if !hasValue(ids) {
		return nil, nil, fmt.Errorf("pk value not provided")
	}

    sql := "DELETE FROM " + self.CurrentTable
	where, values := self.singleCondition(ids, "", extra...)
	if where != "" {
        sql += "\nWHERE " + where
    } else {
        return nil, nil, fmt.Errorf("delete whole table is not supported")
    }
	dbi := &DBI{DB:db}
    return extra, self.Nextpages, dbi.DoSQLContext(ctx, sql, values...)
}
