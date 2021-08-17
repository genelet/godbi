package godbi

import (
	"context"
	"database/sql"
	"fmt"
)

type Edit struct {
	Capability
	Joins    []*Join             `json:"joins,omitempty" hcl:"join,block"`
	Rename   map[string][]string `json:"rename" hcl:"rename"`
	Fields   []string            `json:"fields,omitempty" hcl:"fields"`
}

func (self *Edit) filterPars() (string, []interface{}, string) {
    shorts := make(map[string][2]string)
	for k, v := range self.Rename {
		if self.Fields == nil {
			shorts[k] = [2]string{v[0], v[1]}
		} else {
			if grep(self.Fields, k) {
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

func (self *Edit) Run(db *sql.DB, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
	return self.RunContext(context.Background(), db, ARGS, extra...)
}

func (self *Edit) RunContext(ctx context.Context, db *sql.DB, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
	sql, labels, table := self.filterPars()

    ids := self.getIdVal(ARGS, extra...)
    if !hasValue(ids) {
        return nil, fmt.Errorf("pk value not provided")
    }

    where, extraValues := self.singleCondition(ids, table, extra...)
    if where != "" {
        sql += "\nWHERE " + where
    }

	lists := make([]map[string]interface{}, 0)
	dbi := &DBI{DB:db}
    err := dbi.SelectSQLContext(ctx, &lists, sql, labels, extraValues...)
	if err != nil { return nil, err }
	return lists, nil
}
