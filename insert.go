package godbi

import (
	"context"
	"database/sql"
	"fmt"
)

type Insert struct {
	Action
}

// Run inserts a row using data passed in ARGS. Any value defined
// in 'extra' will override that key in ARGS.
//
func (self *Insert) RunAction(db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
	return self.RunActionContext(context.Background(), db, t, ARGS, extra...)
}

// InsertContext inserts a row using data passed in ARGS. Any value defined
// in 'extra' will override that key in ARGS.
//
func (self *Insert) RunActionContext(ctx context.Context, db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
	err := t.checkNull(ARGS, extra...)
	if err != nil {
		return nil, err
	}

	fieldValues := t.getFv(ARGS)
	if hasValue(extra) && hasValue(extra[0]) {
		for key, value := range extra[0] {
			for _, col := range t.Columns {
				if col.ColumnName == key {
					fieldValues[key] = value
					break
				}
			}
		}
	}
	if fieldValues == nil || len(fieldValues) == 0 {
		return nil, fmt.Errorf("no data to insert")
	}

	autoID, err := t.insertHashContext(ctx, db, fieldValues)
	if err != nil {
		return nil, err
	}

	if t.IdAuto != "" {
		fieldValues[t.IdAuto] = autoID
	}

	return fromFv(fieldValues), nil
}
