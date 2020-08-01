package godbi

import (
	"database/sql"
	"net/url"
	"strings"
	"github.com/golang/glog"
)

// DBI is an abstract database interface
// Db: the generic SQL handler.
// LastId: the last auto id, if any in the table
// Affected: number of row affected after each operation.
type DBI struct {
	Db        *sql.DB
	LastId    int64
	Affected  int64
}

// ExecSQL is the same as the generic SQL's Exec, plus adding
// the affected number of rows into Affected
func (self *DBI) ExecSQL(str string, args ...interface{}) error {
	res, err := self.Db.Exec(str, args...)
	if err != nil {
		return err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	self.LastId = lastID
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	self.Affected = affected

	return nil
}

// DoSQL is the same as ExecSQL, except for using prepared statement,
// which is safe for concurrent use use by multiple goroutines.
func (self *DBI) DoSQL(str string, args ...interface{}) error {
	sth, err := self.Db.Prepare(str)
	if err != nil {
		return err
	}
	res, err := sth.Exec(args...)
	if err != nil {
		return err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	self.LastId = lastID
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	self.Affected = affected

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
		res, err = sth.Exec(once...)
		if err != nil {
			return err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		self.Affected += affected
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	self.LastId = lastID

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

// QuerySQLTypeLabel selects rows and put them into lists, an array of maps.
// It uses the given data types defined in types_labels.
// and the keys in the maps uses the given name defined in selectLabels.
func (self *DBI) QuerySQLTypeLabel(lists *[]map[string]interface{}, typeLabels []string, selectLabels []string, str string, args ...interface{}) error {
	rows, err := self.Db.Query(str, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return self.pickup(rows, lists, typeLabels, selectLabels, str)
}

// SelectSQL is the same as QuerySQL excepts it uses a prepared statement.
func (self *DBI) SelectSQL(lists *[]map[string]interface{}, str string, args ...interface{}) error {
	return self.SelectSQLTypeLabel(lists, nil, nil, str, args...)
}

// SelectSQLType is the same as QuerySQLType excepts it uses a prepared statement.
func (self *DBI) SelectSQLType(lists *[]map[string]interface{}, typeLabels []string, str string, args ...interface{}) error {
	return self.SelectSQLTypeLabel(lists, typeLabels, nil, str, args...)
}

// SelectSQLLabel is the same as QuerySQLLabel excepts it uses a prepared statement.
func (self *DBI) SelectSQLLabel(lists *[]map[string]interface{}, selectLabels []string, str string, args ...interface{}) error {
	return self.SelectSQLTypeLabel(lists, nil, selectLabels, str, args...)
}

// SelectSQLTypeLabel is the same as QuerySQLTypeLabel excepts it uses a prepared statement.
func (self *DBI) SelectSQLTypeLabel(lists *[]map[string]interface{}, typeLabels []string, selectLabels []string, str string, args ...interface{}) error {
glog.Infof("%s", str)
glog.Infof("%v", typeLabels)
glog.Infof("%v", selectLabels)
glog.Infof("%v", args)
	sth, err := self.Db.Prepare(str)
	if err != nil {
		return err
	}
	defer sth.Close()
	rows, err := sth.Query(args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return self.pickup(rows, lists, typeLabels, selectLabels, str)
}

func (self *DBI) pickup(rows *sql.Rows, lists *[]map[string]interface{}, typeLabels []string, selectLabels []string, str string) error {
	var err error
	if selectLabels == nil {
		if selectLabels, err = rows.Columns(); err != nil {
			return err
		}
	}

	isType := false
	if typeLabels != nil {
		isType = true
	}
	names := make([]interface{}, len(selectLabels))
	x := make([]interface{}, len(selectLabels))
	for i := range selectLabels {
		if isType {
			switch typeLabels[i] {
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
		for j, v := range selectLabels {
			if isType {
				switch typeLabels[j] {
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

// GetSQLLabel returns one row and assign them as the values in the res map.
func (self *DBI) GetSQLLabel(res map[string]interface{}, sql string, selectLabels []string, args ...interface{}) error {
	lists := make([]map[string]interface{}, 0)
	if err := self.SelectSQLLabel(&lists, selectLabels, sql, args...); err != nil {
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

// GetArgs returns one row and assign them as the values in ARGS as url.Values
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

// DoProc runs the stored procedure 'proc_name' with input parameters 'args'
// the output is assigned to 'hash' using 'names' as the keys
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

// SelectProc retrieves the results in 'lists', an array of rows,
// from the stored procedure 'proc_name' with input parameters 'args'.
// Each row is represented as a string to interface{} map.
func (self *DBI) SelectProc(lists *[]map[string]interface{}, proc_name string, args ...interface{}) error {
	return self.SelectDoProcLabel(lists, nil, nil, proc_name, nil, args...)
}

// SelectProcLabel retrieves the results in 'lists', an array of rows,
// from the stored procedure 'proc_name' with input parameters 'args'.
// Each row is represented as a labeled string to interface{} map.
func (self *DBI) SelectProcLabel(lists *[]map[string]interface{}, proc_name string, selectLabels []string, args ...interface{}) error {
	return self.SelectDoProcLabel(lists, nil, nil, proc_name, selectLabels, args...)
}

// SelectDoProc retrieves the results in 'lists', an array of rows,
// from the stored procedure 'proc_name' with input parameters 'args'.
// The outputs are saved in hash using keys 'names'.
// Each row is represented as a string to interface{} map.
func (self *DBI) SelectDoProc(lists *[]map[string]interface{}, hash map[string]interface{}, names []string, proc_name string, args ...interface{}) error {
	return self.SelectDoProcLabel(lists, hash, names, proc_name, nil, args...)
}

// SelectDoProc retrieves the results in 'lists', an array of rows,
// from the stored procedure 'proc_name' with input parameters 'args'.
// The outputs are saved in hash using keys 'names'.
// Each row is represented as a labeled string to interface{} map.
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
