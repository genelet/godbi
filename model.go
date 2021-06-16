package godbi

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Model works on table's CRUD in web applications.
//
type Model struct {
	DBI
	Table

	// Updated: for Insupd only, indicating if the row is updated or new
	Updated bool

	// aARGS: the input data received by the web request
	aARGS map[string]interface{}
	// aLISTS: output data as slice of map, which represents a table row
	aLISTS []map[string]interface{}
}

// NewModel creates a new Model struct from json file 'filename'
// You should use SetDB to assign a database handle and
// SetArgs to set input data, a map[string]interface{} to make it working
//
func NewModel(filename string) (*Model, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	table, err := newTable(content)
	if err != nil {
		return nil, err
	}
	model := &Model{Table:*table}
	model.SetArgs(nil)
	return model, nil
}

func (self *Model) GetLists() []map[string]interface{} {
	return self.aLISTS
}

func (self *Model) NonePass() []string {
	return []string{self.Fields, self.Sortby, self.Sortreverse, self.Rowcount, self.Totalno, self.Pageno, self.Maxpageno}
}

// SetArgs sets input data
func (self *Model) SetArgs(args map[string]interface{}) {
	self.aARGS = args
}

// SetDB sets the DB handle
func (self *Model) SetDB(db *sql.DB) {
	self.DB = db
	self.aLISTS = make([]map[string]interface{}, 0)
}

func (self *Model) RunAction(action string, extra ...map[string]interface{}) ([]map[string]interface{}, map[string]interface{}, []*Page, error) {
	var err error
	switch action {
	case "topics", "LIST":
		err = self.Topics(extra...)
	case "edit", "GET":
		err = self.Edit(extra...)
	case "insert", "POST":
		err = self.Insert(extra...)
	case "update", "PUT":
		err = self.Update(extra...)
	case "insupd", "PATCH":
		err = self.Insupd(extra...)
	case "delete", "DELETE":
		err = self.Delete(extra...)
	default:
        return nil, nil, nil, errors.New("action not found in graph")
	}

	if err != nil {
        return nil, nil, nil, err
	}

	modelArgs := make(map[string]interface{})
	nones := self.NonePass()
	for k, v := range self.aARGS {
		if grep(nones, k) {
			continue
		}
		modelArgs[k] = v
	}

	var nextpages []*Page
	if hasValue(self.Nextpages) {
		if nps, ok := self.Nextpages[action]; ok {
			nextpages = nps
		}
	}

	return self.aLISTS, modelArgs, nextpages, nil
}

func (self *Model) filteredFields(pars []string) []string {
	ARGS := self.aARGS
	fields, ok := ARGS[self.Fields]
	if !ok {
		return pars
	}
	out := make([]string, 0)
	for _, field := range fields.([]string) {
		for _, v := range pars {
			if field == v {
				out = append(out, v)
				break
			}
		}
	}
	return out
}

func (self *Model) getFv(pars []string) map[string]interface{} {
	ARGS := self.aARGS
	fieldValues := make(map[string]interface{})
	for _, f := range self.filteredFields(pars) {
		if v, ok := ARGS[f]; ok {
			fieldValues[f] = v
		}
	}
	return fieldValues
}

func (self *Model) getIdVal(extra ...map[string]interface{}) []interface{} {
	if hasValue(self.CurrentKeys) {
		if hasValue(extra) {
			return self.properValues(self.CurrentKeys, extra[0])
		}
		return self.properValues(self.CurrentKeys, nil)
	}
	if hasValue(extra) {
		return []interface{}{self.properValue(self.CurrentKey, extra[0])}
	}
	return []interface{}{self.properValue(self.CurrentKey, nil)}
}

// Topics selects many rows, optionally with restriction defined in 'extra'.
func (self *Model) Topics(extra ...map[string]interface{}) error {
	ARGS := self.aARGS
	totalForce := self.TotalForce // 0 means no total calculation
	if totalForce != 0 && ARGS[self.Rowcount] != nil && ARGS[self.Pageno] == nil {
		nt := 0
		if totalForce < -1 { // take the absolute as the total number
			nt = int(math.Abs(float64(totalForce)))
		} else if totalForce == -1 || ARGS[self.Totalno] == nil { // optionally cal
			if err := self.totalHash(&nt, extra...); err != nil {
				return err
			}
		} else {
			ntInterface := ARGS[self.Totalno]
			switch v := ntInterface.(type) {
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
		nrInterface := ARGS[self.Rowcount]
		switch v := nrInterface.(type) {
		case int:
			nr = v
		case string:
			nr64, err := strconv.ParseInt(v, 10, 32)
			if err != nil { return err }
			nr = int(nr64)
		default:
		}
		maxPageno := (nt-1)/nr + 1
		ARGS[self.Maxpageno] = maxPageno
	}

	hashPars := self.topicsHashPars
    if fields, ok := self.aARGS[self.Fields]; ok {
        hashPars = filterPars(hashPars, fields.([]string))
    }

	self.aLISTS = make([]map[string]interface{}, 0)
	return self.topicsHashOrder(&self.aLISTS, hashPars, self.orderString(), extra...)
}

// Edit selects few rows (usually one) using primary key value in ARGS,
// optionally with restrictions defined in 'extra'.
func (self *Model) Edit(extra ...map[string]interface{}) error {
	val := self.getIdVal(extra...)
	hashPars := self.editHashPars
    if fields, ok := self.aARGS[self.Fields]; ok {
        hashPars = filterPars(hashPars, fields.([]string))
    }
	if !hasValue(val) {
		return errors.New("pk value not provided")
	}

	self.aLISTS = make([]map[string]interface{}, 0)
	return self.editHash(&self.aLISTS, hashPars, val, extra...)
}

// Insert inserts a row using data passed in ARGS. Any value defined
// in 'extra' will override that in ARGS and be used for that column.
func (self *Model) Insert(extra ...map[string]interface{}) error {
	fieldValues := self.getFv(self.InsertPars)
	if hasValue(extra) {
		for key, value := range extra[0] {
			if grep(self.InsertPars, key) {
				fieldValues[key] = value
			}
		}
	}
	if !hasValue(fieldValues) {
		return errors.New("no data to insert")
	}

	self.aLISTS = make([]map[string]interface{}, 0)
	if err := self.insertHash(fieldValues); err != nil {
		return err
	}

	if self.CurrentIDAuto != "" {
		autoID := strconv.FormatInt(self.LastID, 10)
		fieldValues[self.CurrentIDAuto] = autoID
		self.aARGS[self.CurrentIDAuto] = autoID
	}
	self.aLISTS = fromFv(fieldValues)

	return nil
}

// Insupd inserts a new row if it does not exist, or retrieves the old one,
// depending on the unique of the columns defined in InsupdPars.
func (self *Model) Insupd(extra ...map[string]interface{}) error {
	fieldValues := self.getFv(self.InsertPars)
	if hasValue(extra) {
		for key, value := range extra[0] {
			if grep(self.InsertPars, key) {
				fieldValues[key] = value
			}
		}
	}
	if !hasValue(fieldValues) {
		return errors.New("pk value not found")
	}

	uniques := self.InsupdPars
	if !hasValue(uniques) {
		return errors.New("unique key value not found in model")
	}

	if err := self.insupdTable(fieldValues, uniques); err != nil {
		return err
	}

	if self.CurrentIDAuto != "" {
		fieldValues[self.CurrentIDAuto] = strconv.FormatInt(self.LastID, 10)
	}
	self.aLISTS = fromFv(fieldValues)

	return nil
}

// Update updates a row using values defined in ARGS
// depending on the unique of the columns defined in UpdatePars.
// extra is for SQL constrains
func (self *Model) Update(extra ...map[string]interface{}) error {
	val := self.getIdVal(extra...)
	if !hasValue(val) {
		return errors.New("pk value not found")
	}

	fieldValues := self.getFv(self.UpdatePars)
	if !hasValue(fieldValues) {
		return errors.New("no data to update")
	} else if len(fieldValues) == 1 && fieldValues[self.CurrentKey] != nil {
		self.aLISTS = fromFv(fieldValues)
		return nil
	}

	err := self.updateHashNulls(fieldValues, val, self.Empties, extra...)
	if err != nil {
		return err
	}

	if hasValue(self.CurrentKeys) {
		for i, v := range self.CurrentKeys {
			fieldValues[v] = val[i]
		}
	} else {
		fieldValues[self.CurrentKey] = val[0]
	}
	self.aLISTS = fromFv(fieldValues)

	return nil
}

func fromFv(fieldValues map[string]interface{}) []map[string]interface{} {
	return []map[string]interface{}{fieldValues}
}

// Delete deletes a row or multiple rows using the contraint in extra
func (self *Model) Delete(extra ...map[string]interface{}) error {
	if err := self.deleteHash(extra...); err != nil {
		return err
	}

	if hasValue(extra) {
		self.aLISTS = []map[string]interface{}{extra[0]}
	}

	return nil
}

// properValue returns the value of key 'v' from extra.
// In case it does not exist, it tries to get from ARGS.
func (self *Model) properValue(v string, extra map[string]interface{}) interface{} {
	ARGS := self.aARGS
	if !hasValue(extra) {
		return ARGS[v]
	}
	if val, ok := extra[v]; ok {
		return val
	}
	return ARGS[v]
}

// properValues returns the values of multiple keys 'vs' from extra.
// In case it does not exists, it tries to get from ARGS.
func (self *Model) properValues(vs []string, extra map[string]interface{}) []interface{} {
	ARGS := self.aARGS
	outs := make([]interface{}, len(vs))
	if !hasValue(extra) {
		for i, v := range vs {
			outs[i] = ARGS[v]
		}
		return outs
	}
	for i, v := range vs {
		if val, ok := extra[v]; ok {
			outs[i] = val
		} else {
			outs[i] = ARGS[v]
		}
	}
	return outs
}

// properValuesHash is the same as properValues, but resulting in a map.
func (self *Model) properValuesHash(vs []string, extra map[string]interface{}) map[string]interface{} {
	values := self.properValues(vs, extra)
	hash := make(map[string]interface{})
	for i, v := range vs {
		hash[v] = values[i]
	}
	return hash
}

// orderString outputs the ORDER BY string using information in args
func (self *Model) orderString() string {
	ARGS := self.aARGS
	column := ""
	if ARGS[self.Sortby] != nil {
		column = ARGS[self.Sortby].(string)
	} else if hasValue(self.CurrentTables) {
		table := self.CurrentTables[0]
		if table.Sortby != "" {
			column = table.Sortby
		} else {
			name := table.Name
			if table.Alias != "" {
				name = table.Alias
			}
			name += "."
			if hasValue(self.CurrentKeys) {
				column = name + strings.Join(self.CurrentKeys, ", "+name)
			} else {
				column = name + self.CurrentKey
			}
		}
	} else {
		if hasValue(self.CurrentKeys) {
			column = strings.Join(self.CurrentKeys, ", ")
		} else {
			column = self.CurrentKey
		}
	}

	order := "ORDER BY " + column
	if ARGS[self.Sortreverse] != nil {
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
		if _, ok := ARGS[self.Pageno]; !ok {
			ARGS[self.Pageno] = 1
		} else {
			pInterface := ARGS[self.Pageno]
			switch v := pInterface.(type) {
			case int:
				pageno = v
			case string:
				pageno, _ = strconv.Atoi(v)
			default:
			}
		}
		order += " LIMIT " + strconv.Itoa(rowcount) + " OFFSET " + strconv.Itoa((pageno-1)*rowcount)
	}

	matched, err := regexp.MatchString("[;'\"]", order)
	if err != nil || matched {
		return ""
	}
	return order
}
