package godbi

import (
	"context"
	"database/sql"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type Topics struct {
	Action
	Joins  []*Join     `json:"joins,omitempty" hcl:"join,block"`
	Rename []*Col      `json:"rename" hcl:"rename"`
	FIELDS string      `json:"fields,omitempty" hcl:"fields"`

	TotalForce  int    `json:"total_force,omitempty" hcl:"total_force,optional"`
	MAXPAGENO   string `json:"maxpageno,omitempty" hcl:"maxpageno,optional"`
	TOTALNO     string `json:"totalno,omitempty" hcl:"totalno,optional"`
	ROWCOUNT    string `json:"rawcount,omitempty" hcl:"rawcount,optional"`
	PAGENO      string `json:"pageno,omitempty" hcl:"pageno,optional"`
	SORTBY      string `json:"sortby,omitempty" hcl:"sortby,optional"`
	SORTREVERSE string `json:"sortreverse,omitempty" hcl:"sortreverse,optional"`
}

func (self *Topics) defaultNames() []string {
	if self.FIELDS == "" {
		self.FIELDS = "fields"
	}
	if self.SORTBY == "" {
		self.SORTBY = "sortby"
	}
	if self.SORTREVERSE == "" {
		self.SORTREVERSE = "sortreverse"
	}
	if self.ROWCOUNT == "" {
		self.ROWCOUNT = "rowcount"
	}
	if self.PAGENO == "" {
		self.PAGENO = "pageno"
	}
	if self.TOTALNO == "" {
		self.TOTALNO = "totalno"
	}
	if self.MAXPAGENO == "" {
		self.MAXPAGENO = "maxpageno"
	}
	return []string{self.FIELDS, self.SORTBY, self.SORTREVERSE, self.ROWCOUNT, self.PAGENO, self.TOTALNO, self.MAXPAGENO}
}

// orderString outputs the ORDER BY string using information in args
func (self *Topics) orderString(t *Table, ARGS map[string]interface{}) string {
	nameSortby := self.SORTBY
	nameSortreverse := self.SORTREVERSE
	nameRowcount := self.ROWCOUNT
	namePageno := self.PAGENO

	column := ""
	if ARGS[nameSortby] != nil {
		column = ARGS[nameSortby].(string)
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
			column = name + strings.Join(t.Pks, ", "+name)
		}
	} else {
		column = strings.Join(t.Pks, ", ")
	}

	order := "ORDER BY " + column
	if _, ok := ARGS[nameSortreverse]; ok {
		order += " DESC"
	}
	if rowInterface, ok := ARGS[nameRowcount]; ok {
		rowcount := 0
		switch v := rowInterface.(type) {
		case int:
			rowcount = v
		case string:
			rowcount, _ = strconv.Atoi(v)
		default:
		}
		pageno := 1
		if pnInterface, ok := ARGS[namePageno]; ok {
			switch v := pnInterface.(type) {
			case int:
				pageno = v
			case string:
				pageno, _ = strconv.Atoi(v)
			default:
			}
		} else {
			ARGS[namePageno] = 1
		}
		order += " LIMIT " + strconv.Itoa(rowcount) + " OFFSET " + strconv.Itoa((pageno-1)*rowcount)
	}

	matched, err := regexp.MatchString("[;'\"]", order)
	if err != nil || matched {
		return ""
	}
	return order
}

func (self *Topics) pagination(ctx context.Context, db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...interface{}) error {
	nameTotalno := self.TOTALNO
	nameRowcount := self.ROWCOUNT
	namePageno := self.PAGENO
	nameMaxpageno := self.MAXPAGENO

	totalForce := self.TotalForce // 0 means no total calculation
	if totalForce == 0 || ARGS[nameRowcount] == nil || ARGS[namePageno] != nil {
		return nil
	}

	nt := 0
	if totalForce < -1 { // take the absolute as the total number
		nt = int(math.Abs(float64(totalForce)))
	} else if totalForce == -1 || ARGS[nameTotalno] == nil { // optional
		if err := t.totalHashContext(ctx, db, &nt, extra...); err != nil {
			return err
		}
	} else {
		switch v := ARGS[nameTotalno].(type) {
		case int:
			nt = v
		case string:
			nt64, err := strconv.ParseInt(v, 10, 32)
			if err != nil {
				return err
			}
			nt = int(nt64)
		default:
		}
	}

	ARGS[nameTotalno] = nt
	nr := 0
	switch v := ARGS[nameRowcount].(type) {
	case int:
		nr = v
	case string:
		nr64, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return err
		}
		nr = int(nr64)
	default:
	}
	ARGS[nameMaxpageno] = (nt-1)/nr + 1
	return nil
}

func (self *Topics) RunAction(db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...interface{}) ([]map[string]interface{}, []*Edge, error) {
	return self.RunActionContext(context.Background(), db, t, ARGS, extra...)
}

func (self *Topics) RunActionContext(ctx context.Context, db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...interface{}) ([]map[string]interface{}, []*Edge, error) {
	err := self.checkNull(ARGS, extra...)
	if err != nil {
		return nil, nil, err
	}

	self.defaultNames()
	sql, labels, table := self.filterPars(t.TableName, ARGS, self.Rename, self.FIELDS, self.Joins)
	err = self.pagination(ctx, db, t, ARGS, extra...)
	if err != nil {
		return nil, nil, err
	}
	order := self.orderString(t, ARGS)

	dbi := &DBI{DB: db}
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
		if err != nil {
			return nil, nil, err
		}
		return lists, self.Nextpages, nil
	}

	if order != "" {
		sql += "\n" + order
	}

	err = dbi.SelectSQLContext(ctx, &lists, sql, labels)
	if err != nil {
		return nil, nil, err
	}
	return lists, self.Nextpages, nil
}
