package godbi

import (
	"context"
	"database/sql"
	"fmt"
)

type Update struct {
	Insert
	Empties    []string      `json:"empties,omitempty" hcl:"empties,optional"`
}

func (self *Update) Run(db *sql.DB, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
    return self.RunContext(context.Background(), db, ARGS, extra...)
}

func (self *Update) RunContext(ctx context.Context, db *sql.DB, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
    ids := self.getIdVal(ARGS, extra...)
    if !hasValue(ids) {
        return nil, fmt.Errorf("pk value not found")
    }

    fieldValues := getFv(self.Columns, ARGS, nil)
    if !hasValue(fieldValues) {
        return nil, fmt.Errorf("no data to update")
    } else if len(fieldValues) == 1 && fieldValues[self.Pks[0]] != nil {
        return fromFv(fieldValues), nil
    }

    err := self.updateHashNullsContext(ctx, db, fieldValues, ids, self.Empties, extra...)
    return fromFv(fieldValues), err
}
