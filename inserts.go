package godbi

import (
	"context"
	"database/sql"
	"fmt"
)

type Inserts struct {
	Action
	Columns []string `json:"columns,omitempty" hcl:"columns,optional"`
}

// Run inserts a row using data passed in ARGS. Any value defined
// in 'extra' will override that key in ARGS.
//
func (self *Inserts) RunAction(db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...interface{}) ([]map[string]interface{}, []*Edge, error) {
	return self.RunActionContext(context.Background(), db, t, ARGS, extra...)
}

// InsertContext inserts a row using data passed in ARGS. Any value defined
// in 'extra' will override that key in ARGS.
//
func (self *Inserts) RunActionContext(ctx context.Context, db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...interface{}) ([]map[string]interface{}, []*Edge, error) {
	err := self.checkNull(ARGS, extra...)
	if err != nil {
		return nil, nil, err
	}

	var multi []map[string]interface{}

	fieldValues := getFv(self.Columns, ARGS, nil)
	if hasValue(extra) {
		switch v := extra[0].(type) {
		case []map[string]interface{}:
			for _, u := range v {
				item := make(map[string]interface{})
				for key, value := range fieldValues {
					item[key] = value
				}
				for key, value := range u {
					if grep(self.Columns, key) {
						item[key] = value
					}
				}
				if !hasValue(item) {
					return nil, nil, fmt.Errorf("no data to inserts")
				}
				autoID, err := t.insertHashContext(ctx, db, item)
				if err != nil {
					return nil, nil, err
				}
				if t.IDAuto != "" {
					item[t.IDAuto] = autoID
				}
				multi = append(multi, item)
			}
		default:
		}
	}


	return multi, self.Nextpages, nil
}
