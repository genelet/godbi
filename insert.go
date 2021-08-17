package godbi

import (
	"context"
	"database/sql"
	"fmt"
)

type Insert struct {
	Capability
	Columns    []string      `json:"columns,omitempty" hcl:"columns,optional"`
}

// Run inserts a row using data passed in ARGS. Any value defined
// in 'extra' will override that key in ARGS.
//
func (self *Insert) Run(db *sql.DB, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
    return self.RunContext(context.Background(), db, ARGS, extra...)
}

// InsertContext inserts a row using data passed in ARGS. Any value defined
// in 'extra' will override that key in ARGS.
//
func (self *Insert) RunContext(ctx context.Context, db *sql.DB, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
    fieldValues := getFv(self.Columns, ARGS, nil)
    if hasValue(extra) {
        for key, value := range extra[0] {
            if grep(self.Columns, key) {
                fieldValues[key] = value
            }
        }
    }
    if fieldValues==nil || len(fieldValues)==0 {
        return nil, fmt.Errorf("no data to insert")
    }

	autoID, err := self.insertHashContext(ctx, db, fieldValues)
	if err != nil { return nil, err }

    if self.IDAuto != "" {
        fieldValues[self.IDAuto] = autoID
    }

    return fromFv(fieldValues), nil
}
