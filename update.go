package godbi

import (
	"context"
	"database/sql"
	"fmt"
)

type Update struct {
	Action
	Empties []string `json:"empties,omitempty" hcl:"empties,optional"`
}

func (self *Update) RunAction(db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
	return self.RunActionContext(context.Background(), db, t, ARGS, extra...)
}

func (self *Update) RunActionContext(ctx context.Context, db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
	err := t.checkNull(ARGS, extra...)
	if err != nil {
		return nil, err
	}

	ids := t.getIdVal(ARGS, extra...)
	if !hasValue(ids) {
		return nil, fmt.Errorf("pk value not found")
	}

	fieldValues := t.GetFv(ARGS)
	if !hasValue(fieldValues) {
		return nil, fmt.Errorf("no data to update")
	} else if len(fieldValues) == 1 && fieldValues[t.Pks[0]] != nil {
		return fromFv(fieldValues), nil
	}

	err = t.updateHashNullsContext(ctx, db, fieldValues, ids, self.Empties, extra...)
	return fromFv(fieldValues), err
}
