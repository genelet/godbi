package godbi

import (
	"context"
	"database/sql"
	"strings"
	"strconv"
	"regexp"
	"math"
)

type Topics struct {
	Edit
	TotalForce  int    `json:"total_force,omitempty" hcl:"total_force,optional"`
	Maxpageno   string `json:"maxpageno,omitempty" hcl:"maxpageno,optional"`
    Totalno     string `json:"totalno,omitempty" hcl:"totalno,optional"`
    Rowcount    string `json:"rawcount,omitempty" hcl:"rawcount,optional"`
    Pageno      string `json:"pageno,omitempty" hcl:"pageno,optional"`
    Sortreverse string `json:"sortreverse,omitempty" hcl:"sortreverse,optional"`
    Sortby      string `json:"sortby,omitempty" hcl:"sortby,optional"`
}

// orderString outputs the ORDER BY string using information in args
func (self *Topics) orderString(ARGS map[string]interface{}) string {
    column := ""
    if ARGS[self.Sortby] != nil {
        column = ARGS[self.Sortby].(string)
    } else if hasValue(self.Joins) {
        table := self.Joins[0]
        if table.Sortby != "" {
            column = table.Sortby
        } else {
            name := table.Name
            if table.Alias != "" {
                name = table.Alias
            }
            name += "."
            column = name + strings.Join(self.Pks, ", "+name)
        }
    } else {
        column = strings.Join(self.Pks, ", ")
    }

    order := "ORDER BY " + column
    if _, ok := ARGS[self.Sortreverse]; ok {
        order += " DESC"
    }
    if rowInterface, ok := ARGS[self.Rowcount]; ok {
        rowcount := 0
        switch v := rowInterface.(type) {
        case int:
            rowcount = v
        case string:
            rowcount, _ = strconv.Atoi(v)
        default:
        }
        pageno := 1
        if pnInterface, ok := ARGS[self.Pageno]; ok {
            switch v := pnInterface.(type) {
            case int:
                pageno = v
            case string:
                pageno, _ = strconv.Atoi(v)
            default:
            }
        } else {
            ARGS[self.Pageno] = 1
		}
        order += " LIMIT " + strconv.Itoa(rowcount) + " OFFSET " + strconv.Itoa((pageno-1)*rowcount)
    }

    matched, err := regexp.MatchString("[;'\"]", order)
    if err != nil || matched {
        return ""
    }
    return order
}

func (self *Topics) pagination(ctx context.Context, db *sql.DB, ARGS map[string]interface{}, extra ...map[string]interface{}) error {
	totalForce := self.TotalForce // 0 means no total calculation
	if totalForce == 0 || ARGS[self.Rowcount] == nil || ARGS[self.Pageno] != nil {
		return nil
	}

	nt := 0
	if totalForce < -1 { // take the absolute as the total number
		nt = int(math.Abs(float64(totalForce)))
	} else if totalForce == -1 || ARGS[self.Totalno] == nil { // optional
		if err := self.totalHashContext(ctx, db, &nt, extra...); err != nil {
			return err
		}
	} else {
		switch v := ARGS[self.Totalno].(type) {
		case int:
			nt = v
		case string:
			nt64, err := strconv.ParseInt(v, 10, 32)
			if err != nil { return err }
			nt = int(nt64)
		default:
		}
	}

	ARGS[self.Totalno] = nt
	nr := 0
	switch v := ARGS[self.Rowcount].(type) {
	case int:
		nr = v
	case string:
		nr64, err := strconv.ParseInt(v, 10, 32)
		if err != nil { return err }
		nr = int(nr64)
	default:
	}
	ARGS[self.Maxpageno] = (nt-1)/nr + 1
	return nil
}

func (self *Topics) Run(db *sql.DB, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
	return self.RunContext(context.Background(), db, ARGS, extra...)
}

func (self *Topics) RunContext(ctx context.Context, db *sql.DB, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
	sql, labels, table := self.filterPars()
	err := self.pagination(ctx, db, ARGS, extra...)
	if err != nil { return nil, err }
	order := self.orderString(ARGS)

	dbi := &DBI{DB:db}
	lists := make([]map[string]interface{}, 0)
	if hasValue(extra) && hasValue(extra[0]) {
        where, values := selectCondition(extra[0], table)
        if where != "" {
            sql += "\nWHERE " + where
        }
        if order != "" {
            sql += "\n" + order
        }
		err = dbi.SelectSQLContext(ctx, &lists, sql, labels, values...)
		if err != nil { return nil, err }
		return lists, nil
    }

    if order != "" {
        sql += "\n" + order
    }

    err = dbi.SelectSQLContext(ctx, &lists, sql, labels)
	if err != nil { return nil, err }
	return lists, nil
}
