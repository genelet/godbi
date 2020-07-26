package godbi

import (
	"database/sql"
	"net/url"
	"strings"
//	"github.com/golang/glog"
)

// DBI is an abstract database interface
// Db: the generic SQL handler.
// LastId: the last auto id, if any in the table
// Affected: number of row affected after each operation.
type DBI struct {
	Db        *sql.DB
	NeedQuote bool
	LastId    int64
	Affected  int64
}

// Quote escapes string to be used safely in placeholder.
// The SQL functions in the package have already quoted so
// you should not call this again in using them.
func Quote(v interface{}) interface{} {
	switch v.(type) {
	case string:
		str := v.(string)
		str = strings.Trim(str, `'"`)
		str = strings.Replace(str, `'`, `\'`, -1)
		str = strings.Replace(str, `;`, `\;`, -1)
		return `'` + str + `'`
	default:
	}
	return v
}

func (self *DBI) quotes(args []interface{}) []interface{} {
	if args == nil || len(args) == 0 {
		return nil
	}
	if self.NeedQuote == false {
		return args
	}
	new_args := make([]interface{}, 0)
	for _, v := range args {
		new_args = append(new_args, Quote(v))
	}
	return new_args
}

// ExecSQL is the same as the generic SQL's Exec, plus adding
// the affected number of rows into Affected
func (self *DBI) ExecSQL(str string, args ...interface{}) error {
	res, err := self.Db.Exec(str, self.quotes(args)...)
	if err != nil {
		return err
	}

	if lastID, err := res.LastInsertId(); err != nil {
		return err
	} else {
		self.LastId = lastID
	}
	if affected, err := res.RowsAffected(); err != nil {
		return err
	} else {
		self.Affected = affected
	}

	return nil
}

// DoSQL is the same as ExecSQL, except for using prepared statement,
// which is safe for concurrent use use by multiple goroutines.
func (self *DBI) DoSQL(str string, args ...interface{}) error {
	sth, err := self.Db.Prepare(str)
	if err != nil {
		return err
	}
	res, err := sth.Exec(self.quotes(args)...)
	if err != nil {
		return err
	}

	if lastID, err := res.LastInsertId(); err != nil {
		return err
	} else {
		self.LastId = lastID
	}
	if affected, err := res.RowsAffected(); err != nil {
		return err
	} else {
		self.Affected = affected
	}

    sth.Close()
    return nil
}

// DoSQLs adds multiple rows at once, each of the rows is a slice
func (self *DBI) DoSQLs(str string, args ...[]interface{}) error {
	n := len(args)
	if n == 0 {
		return self.DoSQL(str)
	} else if n == 1 {
		return self.DoSQL(str, args[0]...)
	}

	sth, err := self.Db.Prepare(str)
	if err != nil {
		return err
	}

	var res sql.Result
	for _, once := range args {
		res, err = sth.Exec(self.quotes(once)...)
		if err != nil {
			return err
		}
		if affected, err := res.RowsAffected(); err != nil {
			return err
		} else {
			self.Affected += affected
		}
	}
	if lastID, err := res.LastInsertId(); err != nil {
		return err
	} else {
		self.LastId = lastID
	}

    sth.Close()
    return nil
}

// QuerySQL selects rows and put them into lists, an array of maps.
// It lets the generic SQL class to decides rows' data types.
func (self *DBI) QuerySQL(lists *[]map[string]interface{}, str string, args ...interface{}) error {
	return self.QuerySQLTypeLabel(lists, nil, nil, str, args...)
}

// QuerySQLType selects rows and put them into lists, an array of maps.
// It uses the give data types defined in types.
func (self *DBI) QuerySQLType(lists *[]map[string]interface{}, types []string, str string, args ...interface{}) error {
	return self.QuerySQLTypeLabel(lists, types, nil, str, args...)
}

// QuerySQLLabel selects rows and put them into lists, an array of maps.
// The keys in the maps uses the given name defined in labels.
func (self *DBI) QuerySQLLabel(lists *[]map[string]interface{}, labels []string, str string, args ...interface{}) error {
	return self.QuerySQLTypeLabel(lists, nil, labels, str, args...)
}

// QuerySQLLabel selects rows and put them into lists, an array of maps.
// It uses the given data types defined in types_labels.
// and the keys in the maps uses the given name defined in select_labels.
func (self *DBI) QuerySQLTypeLabel(lists *[]map[string]interface{}, type_labels []string, select_labels []string, str string, args ...interface{}) error {
	rows, err := self.Db.Query(str, self.quotes(args)...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return self.pickup(rows, lists, type_labels, select_labels, str)
}

// SelectSQL is the same as QuerySQL excepts it uses a prepared statement.
func (self *DBI) SelectSQL(lists *[]map[string]interface{}, str string, args ...interface{}) error {
	return self.SelectSQLTypeLabel(lists, nil, nil, str, args...)
}

// SelectSQLType is the same as QuerySQLType excepts it uses a prepared statement.
func (self *DBI) SelectSQLType(lists *[]map[string]interface{}, type_labels []string, str string, args ...interface{}) error {
	return self.SelectSQLTypeLabel(lists, type_labels, nil, str, args...)
}

// SelectSQLLable is the same as QuerySQLLabel excepts it uses a prepared statement.
func (self *DBI) SelectSQLLabel(lists *[]map[string]interface{}, select_labels []string, str string, args ...interface{}) error {
	return self.SelectSQLTypeLabel(lists, nil, select_labels, str, args...)
}

// SelectSQLTypeLabel is the same as QuerySQLTypeLabel excepts it uses a prepared statement.
func (self *DBI) SelectSQLTypeLabel(lists *[]map[string]interface{}, type_labels []string, select_labels []string, str string, args ...interface{}) error {
	sth, err := self.Db.Prepare(str)
	if err != nil {
		return err
	}
	defer sth.Close()
	rows, err := sth.Query(self.quotes(args)...)
	if err != nil {
        return err
	}
	defer rows.Close()

	return self.pickup(rows, lists, type_labels, select_labels, str)
}

func (self *DBI) pickup(rows *sql.Rows, lists *[]map[string]interface{}, type_labels []string, select_labels []string, str string) error {
	var err error
	if select_labels == nil {
		if select_labels, err = rows.Columns(); err != nil {
			return err
		}
	}

	isType := false
	if type_labels != nil {
		isType = true
	}
	names := make([]interface{}, len(select_labels))
	x := make([]interface{}, len(select_labels))
	for i := range select_labels {
		if isType {
			switch type_labels[i] {
			case "int", "int8", "int16", "int32", "uint", "uint8", "uint16", "uint32", "int64":
				x[i] = new(sql.NullInt64)
			case "float32", "float64":
				x[i] = new(sql.NullFloat64)
			case "bool":
				x[i] = new(sql.NullBool)
			case "string":
				x[i] = new(sql.NullString)
			default:
			}
		} else {
			x[i] = &names[i]
		}
	}

	for rows.Next() {
		if err = rows.Scan(x...); err != nil {
			return err
		}
		res := make(map[string]interface{})
		for j, v := range select_labels {
			if isType {
				switch type_labels[j] {
				case "int":
					x := x[j].(*sql.NullInt64)
					if x.Valid {
						res[v] = int(x.Int64)
					}
				case "int8":
					x := x[j].(*sql.NullInt64)
					if x.Valid {
						res[v] = int8(x.Int64)
					}
				case "int16":
					x := x[j].(*sql.NullInt64)
					if x.Valid {
						res[v] = int16(x.Int64)
					}
				case "int32":
					x := x[j].(*sql.NullInt64)
					if x.Valid {
						res[v] = int32(x.Int64)
					}
				case "uint":
					x := x[j].(*sql.NullInt64)
					if x.Valid {
						res[v] = uint(x.Int64)
					}
				case "uint8":
					x := x[j].(*sql.NullInt64)
					if x.Valid {
						res[v] = uint8(x.Int64)
					}
				case "uint16":
					x := x[j].(*sql.NullInt64)
					if x.Valid {
						res[v] = uint16(x.Int64)
					}
				case "uint32":
					x := x[j].(*sql.NullInt64)
					if x.Valid {
						res[v] = uint32(x.Int64)
					}
				case "int64":
					x := x[j].(*sql.NullInt64)
					if x.Valid {
						res[v] = x.Int64
					}
				case "float32":
					x := x[j].(*sql.NullFloat64)
					if x.Valid {
						res[v] = float32(x.Float64)
					}
				case "float64":
					x := x[j].(*sql.NullFloat64)
					if x.Valid {
						res[v] = x.Float64
					}
				case "bool":
					x := x[j].(*sql.NullBool)
					if x.Valid {
						res[v] = x.Bool
					}
				case "string":
					x := x[j].(*sql.NullString)
					if x.Valid {
						res[v] = strings.TrimRight(x.String, "\x00")
					}
				default:
				}
			} else {
				name := names[j]
				if name != nil {
					switch val := name.(type) {
					case []uint8:
						res[v] = strings.TrimRight(string(val), "\x00")
					case string:
						res[v] = strings.TrimRight(val, "\x00")
					default:
						res[v] = name
					}
				}
			}
		}
		*lists = append(*lists, res)
	}
	if err := rows.Err(); err != nil && err != sql.ErrNoRows {
		return err
	}
	return nil
}

func (self *DBI) GetSQLLabel(res map[string]interface{}, sql string, select_labels []string, args ...interface{}) error {
	lists := make([]map[string]interface{},0)
	if err := self.SelectSQLLabel(&lists, select_labels, sql, args...); err != nil {
		return err
	}
	if len(lists) >= 1 {
		for k, v := range lists[0] {
			if v != nil {
				res[k] = v
			}
		}
	}
	return nil
}

func (self *DBI) GetArgs(ARGS url.Values, sql string, args ...interface{}) error {
	lists := make([]map[string]interface{}, 0)
	if err := self.SelectSQL(&lists, sql, args...); err != nil {
		return err
	}
	if len(lists) >= 1 {
		for k, v := range lists[0] {
			if v != nil {
				ARGS.Set(k, Interface2String(v))
			}
		}
	}
	return nil
}

func (self *DBI) DoProc(hash map[string]interface{}, names []string, proc_name string, args ...interface{}) error {
	n := len(args)
	strQ := strings.Join(strings.Split(strings.Repeat("?", n), ""), ",")
	str := "CALL " + proc_name + "(" + strQ
	strN := "@" + strings.Join(names, ",@")
	if names != nil {
		str += ", " + strN
	}
	str += ")"

	if err := self.DoSQL(str, args...); err != nil {
		return err
	}
	return self.GetSQLLabel(hash, "SELECT "+strN, names)
}

func (self *DBI) SelectProc(lists *[]map[string]interface{}, proc_name string, args ...interface{}) error {
	return self.SelectDoProcLabel(lists, nil, nil, proc_name, nil, args...)
}

func (self *DBI) SelectProcLabel(lists *[]map[string]interface{}, proc_name string, selectLabels []string, args ...interface{}) error {
	return self.SelectDoProcLabel(lists, nil, nil, proc_name, selectLabels, args...)
}

func (self *DBI) SelectDoProc(lists *[]map[string]interface{}, hash map[string]interface{}, names []string, proc_name string, args ...interface{}) error {
	return self.SelectDoProcLabel(lists, hash, names, proc_name, nil, args...)
}

func (self *DBI) SelectDoProcLabel(lists *[]map[string]interface{}, hash map[string]interface{}, names []string, proc_name string, selectLabels []string, args ...interface{}) error {
	n := len(args)
	strQ := strings.Join(strings.Split(strings.Repeat("?", n), ""), ",")
	str := "CALL " + proc_name + "(" + strQ
	strN := "@" + strings.Join(names, ",@")
	if names != nil {
		str += ", " + strN
	}
	str += ")"

	if err := self.SelectSQLLabel(lists, selectLabels, str, args...); err != nil {
		return err
	}
	if hash == nil {
		return nil
	}
	return self.GetSQLLabel(hash, "SELECT "+strN, names)
}
