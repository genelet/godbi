package godbi

import (
	"context"
	"database/sql"
	"fmt"
)

type Edit struct {
	Action
	Joins  []*Join             `json:"joins,omitempty" hcl:"join,block"`
	Rename map[string][]string `json:"rename" hcl:"rename"`
	FIELDS string              `json:"fields,omitempty" hcl:"fields"`
}

func (self *Edit) defaultNames() []string {
	if self.FIELDS == "" {
		self.FIELDS = "fields"
	}
	return []string{self.FIELDS}
}

func (self *Edit) RunAction(db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, []*Edge, error) {
	return self.RunActionContext(context.Background(), db, t, ARGS, extra...)
}

func (self *Edit) RunActionContext(ctx context.Context, db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, []*Edge, error) {
	err := self.checkNull(ARGS)
	if err != nil {
		return nil, nil, err
	}

	self.defaultNames()
	sql, labels, table := self.filterPars(t.CurrentTable, ARGS, self.Rename, self.FIELDS, self.Joins)

	ids := t.getIdVal(ARGS, extra...)
	if !hasValue(ids) {
		return nil, nil, fmt.Errorf("pk value not provided")
	}

	where, extraValues := t.singleCondition(ids, table, extra...)
	if where != "" {
		sql += "\nWHERE " + where
	}

	lists := make([]map[string]interface{}, 0)
	dbi := &DBI{DB: db}
	err = dbi.SelectSQLContext(ctx, &lists, sql, labels, extraValues...)
	if err != nil {
		return nil, nil, err
	}
	return lists, self.Nextpages, nil
}
