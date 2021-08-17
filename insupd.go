package godbi

import (
	"context"
	"database/sql"
	"fmt"
)

type Insupd struct {
	Insert
	Uniques    []string      `json:"uniques,omitempty" hcl:"uniques,optional"`
}

func (self *Insupd) Run(db *sql.DB, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
    return self.RunContext(context.Background(), db, ARGS, extra...)
}

func (self *Insupd) RunContext(ctx context.Context, db *sql.DB, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
    if self.Uniques == nil {
        return nil, fmt.Errorf("unique key not found")
    }

    fieldValues := getFv(self.Columns, ARGS, nil)
    if hasValue(extra) {
        for key, value := range extra[0] {
            if grep(self.Columns, key) {
                fieldValues[key] = value
            }
        }
    }
    if fieldValues == nil || len(fieldValues) == 0 {
        return nil, fmt.Errorf("input not found")
    }

    changed, err := self.insupdTableContext(ctx, db, fieldValues, self.Uniques)
	if err != nil {
        return nil, err
    }

    if self.IDAuto != "" {
        fieldValues[self.IDAuto] = changed
    }

    return fromFv(fieldValues), nil
}
