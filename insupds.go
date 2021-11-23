package godbi

import (
	"context"
	"database/sql"
	"fmt"
)

type Insupds struct {
	Action
	Columns []string `json:"columns,omitempty" hcl:"columns,optional"`
	Uniques []string `json:"uniques,omitempty" hcl:"uniques,optional"`
}

func (self *Insupds) RunAction(db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...interface{}) ([]map[string]interface{}, []*Edge, error) {
	return self.RunActionContext(context.Background(), db, t, ARGS, extra...)
}

func (self *Insupds) RunActionContext(ctx context.Context, db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...interface{}) ([]map[string]interface{}, []*Edge, error) {
	err := self.checkNull(ARGS, extra...)
	if err != nil {
		return nil, nil, err
	}

	if self.Uniques == nil {
		return nil, nil, fmt.Errorf("unique key not found")
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
					return nil, nil, fmt.Errorf("input not found")
				}
				changed, err := t.insupdTableContext(ctx, db, item, self.Uniques)
				if err != nil {
					return nil, nil, err
				}

				if t.IdAuto != "" {
					item[t.IdAuto] = changed
				}
				multi = append(multi, item)
			}
		default:
		}
	}

	return multi, self.Nextpages, nil
}
