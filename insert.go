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
func (self *Insert) RunAction(db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...interface{}) ([]map[string]interface{}, []*Edge, error) {
	return self.RunActionContext(context.Background(), db, t, ARGS, extra...)
}

// InsertContext inserts a row using data passed in ARGS. Any value defined
// in 'extra' will override that key in ARGS.
//
func (self *Insert) RunActionContext(ctx context.Context, db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...interface{}) ([]map[string]interface{}, []*Edge, error) {
	err := self.checkNull(ARGS, extra...)
	if err != nil {
		return nil, nil, err
	}

	fieldValues := getFv(self.Columns, ARGS, nil)
	if hasValue(extra) {
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
		return nil, nil, fmt.Errorf("no data to insert")
	}

	autoID, err := t.insertHashContext(ctx, db, fieldValues)
	if err != nil {
		return nil, nil, err
	}

	if t.IdAuto != "" {
		fieldValues[t.IdAuto] = autoID
	}

	return fromFv(fieldValues), self.Nextpages, nil
}
