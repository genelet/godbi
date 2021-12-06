package godbi

import (
	"context"
	"database/sql"
	"fmt"
)

type Insert struct {
	Action
	Columns []string `json:"columns,omitempty" hcl:"columns,optional"`
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
	err := self.checkNull(ARGS, extra...)
	if err != nil {
		return nil, err
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
