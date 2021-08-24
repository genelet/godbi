package godbi

import (
	"context"
	"database/sql"
	"fmt"
)

type Edit struct {
	Action
	Joins    []*Join             `json:"joins,omitempty" hcl:"join,block"`
	Rename   map[string][]string `json:"rename" hcl:"rename"`
	FIELDS   string              `json:"fields,omitempty" hcl:"fields"`
}

func (self *Edit) defaultNames() []string {
    if self.FIELDS=="" { self.FIELDS = "fields" }
    return []string{self.FIELDS}
}

/*
func (self *Edit) filterPars(ARGS map[string]interface{}) (string, []interface{}, string) {
	var fields []string
	if v, ok := ARGS[self.FIELDS]; ok {
		fields = v.([]string)
	}

    shorts := make(map[string][2]string)
	for k, v := range self.Rename {
		if fields == nil {
			shorts[k] = [2]string{v[0], v[1]}
		} else {
			if grep(fields, k) {
				shorts[k] = [2]string{v[0], v[1]}
			}
		}
	}

	sql, labels := selectType(shorts)
    var table string
    if hasValue(self.Joins) {
        sql = "SELECT " + sql + "\nFROM " + joinString(self.Joins)
        table = self.Joins[0].getAlias()
    } else {
        sql = "SELECT " + sql + "\nFROM " + self.CurrentTable
    }

	return sql, labels, table
}
*/

func (self *Edit) RunAction(db *sql.DB, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, []*Page, error) {
	return self.RunActionContext(context.Background(), db, ARGS, extra...)
}

func (self *Edit) RunActionContext(ctx context.Context, db *sql.DB, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, []*Page, error) {
	err := self.checkNull(ARGS)
    if err != nil { return nil, nil, err }

	self.defaultNames()
	sql, labels, table := self.filterPars(ARGS, self.Rename, self.FIELDS, self.Joins)

    ids := self.getIdVal(ARGS, extra...)
    if !hasValue(ids) {
        return nil, nil, fmt.Errorf("pk value not provided")
    }

    where, extraValues := self.singleCondition(ids, table, extra...)
    if where != "" {
        sql += "\nWHERE " + where
    }

	lists := make([]map[string]interface{}, 0)
	dbi := &DBI{DB:db}
    err = dbi.SelectSQLContext(ctx, &lists, sql, labels, extraValues...)
	if err != nil { return nil, nil, err }
	return lists, self.Nextpages, nil
}
