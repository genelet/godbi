package godbi

import (
	// "github.com/golang/glog"
	"database/sql"
	"net/url"
	"strings"
)

// DBI simply embeds GO's generic SQL handler.
// It adds a set of functions for easier database executionis and queriesr.
//
type DBI struct {
	// Embedding the generic database handle.
	*sql.DB
	// LastId: the last auto id inserted, if the database provides
	LastId int64
	// Affected: the number of rows affected
	Affected int64
}

// ExecSQL excutes a query like 'Exec', and refreshes the LastId and Affected
// If the execution fails, it returns error; otherwise nil.
//
func (self *DBI) ExecSQL(query string, args ...interface{}) error {
	//glog.Infof("godbi SQL statement: %s", query)
	//glog.Infof("godbi input data: %v", args)

	res, err := self.DB.Exec(query, args...)
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

// DoSQL is the same as ExecSQL, except for using a prepared statement,
// which is safe for concurrent use by multiple goroutines.
//
func (self *DBI) DoSQL(query string, args ...interface{}) error {
	//glog.Infof("godbi SQL statement: %s", query)
	//glog.Infof("godbi input data: %v", args)

	sth, err := self.DB.Prepare(query)
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

// DoSQLs inserts multiple rows at once.
// Each row is represented as array and the rows are array of array.
//
func (self *DBI) DoSQLs(query string, args ...[]interface{}) error {
	//glog.Infof("godbi SQL statement: %s", query)
	//glog.Infof("godbi input data: %v", args)

	n := len(args)
	if n == 0 {
		return self.DoSQL(query)
	} else if n == 1 {
		return self.DoSQL(query, args[0]...)
	}

	sth, err := self.DB.Prepare(query)
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

// QuerySQL selects data rows as slice of maps.
// The rows' data types are determined dynamically by the generic handle.
// 'lists' is a reference to slice of maps to receive the quering data.
//
func (self *DBI) QuerySQL(lists *[]map[string]interface{}, query string, args ...interface{}) error {
	return self.QuerySQLTypeLabel(lists, nil, nil, query, args...)
}

// QuerySQL selects data rows as slice of maps.
// The rows' data types are defined in 'types' as an array of string.
// 'lists' is a reference to slice of maps to receive the quering data.
//
func (self *DBI) QuerySQLType(lists *[]map[string]interface{}, typeLabels []string, query string, args ...interface{}) error {
	return self.QuerySQLTypeLabel(lists, typeLabels, nil, query, args...)
}

// QuerySQL selects data rows as slice of maps.
// The rows' data types are determined dynamically by the generic handle.
// 'lists' is a reference to slice of maps to receive the quering data.
// The original SQL column names will be replaced by 'selectLabels'.
//
func (self *DBI) QuerySQLLabel(lists *[]map[string]interface{}, selectLabels []string, query string, args ...interface{}) error {
	return self.QuerySQLTypeLabel(lists, nil, selectLabels, query, args...)
}

//  selects data rows as slice of maps.
// The rows' data types are defined in 'typeLabels' as an array of string.
// 'lists' is a reference to slice of maps to receive the quering data.
// The original SQL column names will be replaced by 'labels'.
//
func (self *DBI) QuerySQLTypeLabel(lists *[]map[string]interface{}, typeLabels []string, selectLabels []string, query string, args ...interface{}) error {
	//glog.Infof("godbi SQL statement: %s", query)
	//glog.Infof("godbi select columns: %v", selectLabels)
	//glog.Infof("godbi column types: %v", typeLabels)
	//glog.Infof("godbi input data: %v", args)

	rows, err := self.DB.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return self.pickup(rows, lists, typeLabels, selectLabels, query)
}

// SelectSQL is the same as QuerySQL excepts it uses a prepared statement.
//
func (self *DBI) SelectSQL(lists *[]map[string]interface{}, query string, args ...interface{}) error {
	return self.SelectSQLTypeLabel(lists, nil, nil, query, args...)
}

// SelectSQLType is the same as QuerySQLType excepts it uses a prepared statement.
//
func (self *DBI) SelectSQLType(lists *[]map[string]interface{}, typeLabels []string, query string, args ...interface{}) error {
	return self.SelectSQLTypeLabel(lists, typeLabels, nil, query, args...)
}

// SelectSQLLabel is the same as QuerySQLLabel excepts it uses a prepared statement.
//
func (self *DBI) SelectSQLLabel(lists *[]map[string]interface{}, selectLabels []string, query string, args ...interface{}) error {
	return self.SelectSQLTypeLabel(lists, nil, selectLabels, query, args...)
}

// SelectSQLTypeLabel is the same as QuerySQLTypeLabel excepts it uses a prepared statement.
//
func (self *DBI) SelectSQLTypeLabel(lists *[]map[string]interface{}, typeLabels []string, selectLabels []string, query string, args ...interface{}) error {
	//glog.Infof("godbi SQL statement: %s", query)
	//glog.Infof("godbi select columns: %v", selectLabels)
	//glog.Infof("godbi column types: %v", typeLabels)
	//glog.Infof("godbi input data: %v", args)

	sth, err := self.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer sth.Close()
	rows, err := sth.Query(args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return self.pickup(rows, lists, typeLabels, selectLabels, query)
}

func (self *DBI) pickup(rows *sql.Rows, lists *[]map[string]interface{}, typeLabels []string, selectLabels []string, query string) error {
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

// GetSQLLabel returns one row as map into 'res'.
// The column names are replaced by 'selectLabels'
//
func (self *DBI) GetSQLLabel(res map[string]interface{}, query string, selectLabels []string, args ...interface{}) error {
	lists := make([]map[string]interface{}, 0)
	if err := self.SelectSQLLabel(&lists, selectLabels, query, args...); err != nil {
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

// GetArgs returns one row as url.Values into 'res', as in web application.
//
func (self *DBI) GetArgs(res url.Values, query string, args ...interface{}) error {
	lists := make([]map[string]interface{}, 0)
	if err := self.SelectSQL(&lists, query, args...); err != nil {
		return err
	}
	if len(lists) >= 1 {
		for k, v := range lists[0] {
			if v != nil {
				res.Set(k, interface2String(v))
			}
		}
	}
	return nil
}

// DoProc runs the stored procedure 'proc_name' using input parameters 'args'.
// The output is returned as map using 'names' as keys.
//
func (self *DBI) DoProc(res map[string]interface{}, names []string, proc_name string, args ...interface{}) error {
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
	return self.GetSQLLabel(res, "SELECT "+strN, names)
}

// SelectProc runs the stored procedure 'proc_name' using input parameters 'args'.
// The query result, 'lists', is received as a slice of maps,
// with data types determined dynamically by the generic SQL handle.
//
func (self *DBI) SelectProc(lists *[]map[string]interface{}, proc_name string, args ...interface{}) error {
	return self.SelectDoProcLabel(lists, nil, nil, proc_name, nil, args...)
}

/*
// SelectProcLabel runs the stored procedure 'proc_name' using input parameters 'args'.
// The query result, 'lists', is received as a slice of maps.
// The keys of the maps are renamed according to 'selectLabels'.
//
func (self *DBI) SelectProcLabel(lists *[]map[string]interface{}, proc_name string, selectLabels []string, args ...interface{}) error {
	return self.SelectDoProcLabel(lists, nil, nil, proc_name, selectLabels, args...)
}
*/

// SelectDoProc runs the stored procedure 'proc_name' using input parameters 'args'.
// The query result, 'lists', is received as a slice of maps,
// with data types are dynamically determined by the generic SQL handle.
// The output is returned as map using 'names' as keys.
//
func (self *DBI) SelectDoProc(lists *[]map[string]interface{}, hash map[string]interface{}, names []string, proc_name string, args ...interface{}) error {
	return self.SelectDoProcLabel(lists, hash, names, proc_name, nil, args...)
}

// SelectDoProcLabel runs the stored procedure 'proc_name' using input parameters 'args'.
// The query result, 'lists', is received as a slice of maps,
// with data types are dynamically determined by the generic SQL handle,
// and the map keys are renamed according to selectLabels.
// The output is returned as map using 'names' as keys.
//
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
