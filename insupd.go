package godbi

import (
	"context"
	"database/sql"
	"fmt"
)

type Insupd struct {
	Action
	Columns []string `json:"columns,omitempty" hcl:"columns,optional"`
	Uniques []string `json:"uniques,omitempty" hcl:"uniques,optional"`
}

func (self *Insupd) RunAction(db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
	return self.RunActionContext(context.Background(), db, t, ARGS, extra...)
}

func (self *Insupd) RunActionContext(ctx context.Context, db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
	err := self.checkNull(ARGS, extra...)
	if err != nil {
		return nil, err
	}

	if self.Uniques == nil {
		return nil, fmt.Errorf("unique key not found")
	}

	fieldValues := getFv(self.Columns, ARGS, nil)
	if hasValue(extra) && hasValue(extra[0]) {
		for key, value := range extra[0] {
			if grep(self.Columns, key) {
				fieldValues[key] = value
			}
		}
	}
	if fieldValues == nil || len(fieldValues) == 0 {
		return nil, fmt.Errorf("input not found")
	}

	changed, err := t.insupdTableContext(ctx, db, fieldValues, self.Uniques)
	if err != nil {
		return nil, err
	}

	if t.IdAuto != "" {
		fieldValues[t.IdAuto] = changed
	}

	return fromFv(fieldValues), nil
}
