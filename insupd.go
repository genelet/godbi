package godbi

import (
	"context"
	"database/sql"
	"fmt"
)

type Insupd struct {
	Action
}

func (self *Insupd) RunAction(db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
	return self.RunActionContext(context.Background(), db, t, ARGS, extra...)
}

func (self *Insupd) RunActionContext(ctx context.Context, db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
	err := t.checkNull(ARGS, extra...)
	if err != nil {
		return nil, err
	}

	fieldValues := t.getFv(ARGS)
	if hasValue(extra) && hasValue(extra[0]) {
		for key, value := range extra[0] {
			for _, col := range t.Rename {
				if col.ColumnName == key {
					fieldValues[key] = value
					break
				}
			}
		}
	}
	if fieldValues == nil || len(fieldValues) == 0 {
		return nil, fmt.Errorf("input not found")
	}

	changed, err := t.insupdTableContext(ctx, db, fieldValues)
	if err != nil {
		return nil, err
	}

	if t.IdAuto != "" {
		fieldValues[t.IdAuto] = changed
	}

	return fromFv(fieldValues), nil
}
