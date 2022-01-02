package godbi

import (
	"context"
	"database/sql"
	"fmt"
)

type Edit struct {
	Action
	Joints []*Joint `json:"joins,omitempty" hcl:"join,block"`
	FIELDS string   `json:"fields,omitempty" hcl:"fields"`
}

func (self *Edit) setDefaultElementNames() []string {
	if self.FIELDS == "" {
		self.FIELDS = "fields"
	}
	return []string{self.FIELDS}
}

func (self *Edit) RunAction(db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
	return self.RunActionContext(context.Background(), db, t, ARGS, extra...)
}

func (self *Edit) RunActionContext(ctx context.Context, db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
	self.setDefaultElementNames()
	sql, labels, table := t.filterPars(ARGS, self.FIELDS, self.Joints)

	ids := t.getIdVal(ARGS, extra...)
	if !hasValue(ids) {
		return nil, fmt.Errorf("pk value not provided")
	}

	where, extraValues := t.singleCondition(ids, table, extra...)
	if where != "" {
		sql += "\nWHERE " + where
	}

	lists := make([]map[string]interface{}, 0)
	dbi := &DBI{DB: db}
	if t.questionNumber == Postgres { sql = questionMarkerNumber(sql) }
	err := dbi.SelectSQLContext(ctx, &lists, sql, labels, extraValues...)
	return lists, err
}
