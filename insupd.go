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

func (self *Insupd) RunAction(db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...interface{}) ([]map[string]interface{}, []*Edge, error) {
	return self.RunActionContext(context.Background(), db, t, ARGS, extra...)
}

func (self *Insupd) RunActionContext(ctx context.Context, db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...interface{}) ([]map[string]interface{}, []*Edge, error) {
	err := self.checkNull(ARGS, extra...)
	if err != nil {
		return nil, nil, err
	}

	if self.Uniques == nil {
		return nil, nil, fmt.Errorf("unique key not found")
	}

	fieldValues := getFv(self.Columns, ARGS, nil)
	if hasValue(extra) && hasValue(extra[0]) {
		switch v := extra[0].(type) {
		case map[string]interface{}:
			for key, value := range v {
				if grep(self.Columns, key) {
					fieldValues[key] = value
				}
			}
		default:
		}
	}
	if fieldValues == nil || len(fieldValues) == 0 {
		return nil, nil, fmt.Errorf("input not found")
	}

	changed, err := t.insupdTableContext(ctx, db, fieldValues, self.Uniques)
	if err != nil {
		return nil, nil, err
	}

	if t.IdAuto != "" {
		fieldValues[t.IdAuto] = changed
	}

	return fromFv(fieldValues), self.Nextpages, nil
}
