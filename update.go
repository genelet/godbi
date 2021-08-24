package godbi

import (
	"context"
	"database/sql"
	"fmt"
)

type Update struct {
	Action
	Columns    []string      `json:"columns,omitempty" hcl:"columns,optional"`
	Empties    []string      `json:"empties,omitempty" hcl:"empties,optional"`
}

func (self *Update) RunAction(db *sql.DB, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, []*Page, error) {
    return self.RunActionContext(context.Background(), db, ARGS, extra...)
}

func (self *Update) RunActionContext(ctx context.Context, db *sql.DB, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, []*Page, error) {
	err := self.checkNull(ARGS)
    if err != nil { return nil, nil, err }

    ids := self.getIdVal(ARGS, extra...)
    if !hasValue(ids) {
        return nil, nil, fmt.Errorf("pk value not found")
    }

    fieldValues := getFv(self.Columns, ARGS, nil)
    if !hasValue(fieldValues) {
        return nil, nil, fmt.Errorf("no data to update")
    } else if len(fieldValues) == 1 && fieldValues[self.Pks[0]] != nil {
        return fromFv(fieldValues), self.Nextpages, nil
    }

    err = self.updateHashNullsContext(ctx, db, fieldValues, ids, self.Empties, extra...)
    return fromFv(fieldValues), self.Nextpages, err
}
