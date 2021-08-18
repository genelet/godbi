package godbi

import (
	"context"
	"database/sql"
	"fmt"
)

type Delete struct {
	Action
}

func (self *Delete) Run(db *sql.DB, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, []*Page, error) {
    return self.RunContext(context.Background(), db, ARGS, extra...)
}

func (self *Delete) RunContext(ctx context.Context, db *sql.DB, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, []*Page, error) {
	err := self.CheckNull(ARGS)
    if err != nil { return nil, nil, err }

    sql := "DELETE FROM " + self.CurrentTable
    if !hasValue(extra) {
        return nil, nil, fmt.Errorf("delete whole table is not supported")
    }
    where, values := selectCondition(extra[0], "")
    if where != "" {
        sql += "\nWHERE " + where
    }
	dbi := &DBI{DB:db}
    return extra, self.Nextpages, dbi.DoSQLContext(ctx, sql, values...)
}
