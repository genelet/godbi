package godbi

import (
	"context"
	"database/sql"
	"fmt"
)

type Update struct {
	Action
	Columns []string `json:"columns,omitempty" hcl:"columns,optional"`
	Empties []string `json:"empties,omitempty" hcl:"empties,optional"`
}

func (self *Update) RunAction(db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...interface{}) ([]map[string]interface{}, []*Edge, error) {
	return self.RunActionContext(context.Background(), db, t, ARGS, extra...)
}

func (self *Update) RunActionContext(ctx context.Context, db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...interface{}) ([]map[string]interface{}, []*Edge, error) {
	err := self.checkNull(ARGS)
	if err != nil {
		return nil, nil, err
	}

	ids := t.getIdVal(ARGS, extra...)
	if !hasValue(ids) {
		return nil, nil, fmt.Errorf("pk value not found")
	}

	fieldValues := getFv(self.Columns, ARGS, nil)
	if !hasValue(fieldValues) {
		return nil, nil, fmt.Errorf("no data to update")
	} else if len(fieldValues) == 1 && fieldValues[t.Pks[0]] != nil {
		return fromFv(fieldValues), self.Nextpages, nil
	}

	err = t.updateHashNullsContext(ctx, db, fieldValues, ids, self.Empties, extra...)
	return fromFv(fieldValues), self.Nextpages, err
}
