package godbi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// Navigate is interface to implement Model
//
type Navigate interface {
	// GetAction: get an action function by name
	GetAction(string) func(...url.Values) error

	// GetLists: get the main data
	GetLists() []map[string]interface{}

	// getArgs: get ARGS; pass "true" for nextpages
	getArgs(...bool) url.Values

	// setArgs: set new input
	SetArgs(url.Values)

	// getNextpages: get the nextpages
	getNextpages(string) []*Page

	// setDB: set SQL handle
	SetDB(*sql.DB)
}

// Model works on table's CRUD in web applications.
//
type Model struct {
	DBI
	Table
	Navigate

	// Actions: map between name and action functions
	Actions map[string]func(...url.Values) error
	// Updated: for Insupd only, indicating if the row is updated or new
	Updated bool

	// aARGS: the input data received by the web request
	aARGS url.Values
	// aLISTS: output data as slice of map, which represents a table row
	aLISTS []map[string]interface{}
}

// NewModel creates a new Model struct from json file 'filename'
// You should use SetDB to assign a database handle and
// SetArgs to set input data, a url.Value, to make it working
//
func NewModel(filename string) (*Model, error) {
	var parsed *Model
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(content, &parsed)
	if err != nil {
		return nil, err
	}

	if parsed.Sortby == "" {
		parsed.Sortby = "sortby"
	}
	if parsed.Sortreverse == "" {
		parsed.Sortreverse = "sortreverse"
	}
	if parsed.Pageno == "" {
		parsed.Pageno = "pageno"
	}
	if parsed.Totalno == "" {
		parsed.Totalno = "totalno"
	}
	if parsed.Rowcount == "" {
		parsed.Rowcount = "rowcount"
	}
	if parsed.Maxpageno == "" {
		parsed.Maxpageno = "maxpage"
	}
	if parsed.Fields == "" {
		parsed.Fields = "fields"
	}
	if parsed.Empties == "" {
		parsed.Empties = "empties"
	}

	return parsed, nil
}

// GetLists get main data as slice of mapped row
func (self *Model) GetLists() []map[string]interface{} {
	return self.aLISTS
}

// GetAction returns action's function
func (self *Model) GetAction(action string) func(...url.Values) error {
	if act, ok := self.Actions[action]; ok {
		return act
	}

	return nil
}

// getArgs returns the input data which may have extra keys added
// pass true will turn off those pagination data
func (self *Model) getArgs(is ...bool) url.Values {
	args := url.Values{}
	for k, v := range self.aARGS {
		if is != nil && is[0] && grep([]string{self.Sortby, self.Sortreverse, self.Rowcount, self.Totalno, self.Pageno, self.Maxpageno}, k) {
			continue
		}
		args[k] = v
	}

	return args
}

// SetArgs sets input data
func (self *Model) SetArgs(args url.Values) {
	self.aARGS = args
}

// getNextpages returns the next pages of an action
func (self *Model) getNextpages(action string) []*Page {
	if !hasValue(self.Nextpages) {
		return nil
	}
	nps, ok := self.Nextpages[action]
	if !ok {
		return nil
	}
	return nps
}

// SetDB sets the DB handle
func (self *Model) SetDB(db *sql.DB) {
	self.DB = db
	self.aLISTS = make([]map[string]interface{}, 0)
}

func (self *Model) filteredFields(pars []string) []string {
	ARGS := self.aARGS
	fields, ok := ARGS[self.Fields]
	if !ok {
		return pars
	}

	out := make([]string, 0)
	for _, field := range fields {
		for _, v := range pars {
			if field == v {
				out = append(out, v)
				break
			}
		}
	}
	return out
}

func (self *Model) getFv(pars []string) url.Values {
	ARGS := self.aARGS
	fieldValues := url.Values{}
	for _, f := range self.filteredFields(pars) {
		if v, ok := ARGS[f]; ok {
			fieldValues[f] = v
		}
	}
	return fieldValues
}

func (self *Model) getIdVal(extra ...url.Values) []interface{} {
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
func (self *Model) Topics(extra ...url.Values) error {
	ARGS := self.aARGS
	totalForce := self.TotalForce // 0 means no total calculation
	if totalForce != 0 && ARGS.Get(self.Rowcount) != "" && (ARGS.Get(self.Pageno) == "" || ARGS.Get(self.Pageno) == "1") {
		nt := int64(0)
		if totalForce < -1 { // take the absolute as the total number
			nt = int64(math.Abs(float64(totalForce)))
		} else if totalForce == -1 || ARGS.Get(self.Totalno) == "" { // optionally cal
			if err := self.totalHash(&nt, extra...); err != nil {
				return err
			}
		} else {
			nt, _ = strconv.ParseInt(ARGS.Get(self.Totalno), 10, 32)
		}
		ARGS.Set(self.Totalno, strconv.FormatInt(nt, 10))
		nr, _ := strconv.ParseInt(ARGS.Get(self.Rowcount), 10, 32)
		maxPageno := int64((nt-1)/nr) + 1
		ARGS.Set(self.Maxpageno, strconv.FormatInt(maxPageno, 10))
	}

	var fields interface{}
	if self.topicsHashPars == nil {
		fields = self.filteredFields(self.TopicsPars)
	} else {
		fields = self.topicsHashPars
	}
	self.aLISTS = make([]map[string]interface{}, 0)
	return self.topicsHashOrder(&self.aLISTS, fields, self.OrderString(), extra...)
}

// Edit selects few rows (usually one) using primary key value in ARGS,
// optionally with restrictions defined in 'extra'.
func (self *Model) Edit(extra ...url.Values) error {
	val := self.getIdVal(extra...)
	fields := self.filteredFields(self.EditPars)
	if !hasValue(fields) {
		return errors.New("pk value not provided")
	}

	self.aLISTS = make([]map[string]interface{}, 0)
	if hasValue(extra) {
		return self.editHash(&self.aLISTS, fields, val, extra[0])
	}
	return self.editHash(&self.aLISTS, fields, val)
}

// Insert inserts a row using data passed in ARGS. Any value defined
// in 'extra' will override that in ARGS and be used for that column.
func (self *Model) Insert(extra ...url.Values) error {
	fieldValues := self.getFv(self.InsertPars)
	if hasValue(extra) {
		for key, value := range extra[0] {
			if grep(self.InsertPars, key) {
				fieldValues.Set(key, value[0])
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
		fieldValues.Set(self.CurrentIDAuto, autoID)
		self.aARGS.Set(self.CurrentIDAuto, autoID)
	}
	self.aLISTS = fromFv(fieldValues)

	return nil
}

// Insupd inserts a new row if it does not exist, or retrieves the old one,
// depending on the unique of the columns defined in InsupdPars.
func (self *Model) Insupd(extra ...url.Values) error {
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
		return errors.New("unique key value not found")
	}

	if err := self.insupdTable(fieldValues, uniques); err != nil {
		return err
	}

	if self.CurrentIDAuto != "" {
		fieldValues.Set(self.CurrentIDAuto, strconv.FormatInt(self.LastID, 10))
	}
	self.aLISTS = fromFv(fieldValues)

	return nil
}

// Update updates a row using values defined in ARGS
// depending on the unique of the columns defined in UpdatePars.
// extra is for SQL constrains
func (self *Model) Update(extra ...url.Values) error {
	val := self.getIdVal(extra...)
	if !hasValue(val) {
		return errors.New("pk value not found")
	}

	fieldValues := self.getFv(self.UpdatePars)
	if !hasValue(fieldValues) {
		return errors.New("no data to update")
	} else if len(fieldValues) == 1 && fieldValues.Get(self.CurrentKey) != "" {
		self.aLISTS = fromFv(fieldValues)
		return nil
	}

	if err := self.updateHashNulls(fieldValues, val, self.aARGS[self.Empties], extra...); err != nil {
		return err
	}

	if hasValue(self.CurrentKeys) {
		for i, v := range self.CurrentKeys {
			fieldValues.Set(v, val[i].(string))
		}
	} else {
		fieldValues.Set(self.CurrentKey, val[0].(string))
	}
	self.aLISTS = fromFv(fieldValues)

	return nil
}

func fromFv(fieldValues url.Values) []map[string]interface{} {
	hash := make(map[string]interface{})
	for k, v := range fieldValues {
		hash[k] = v[0]
	}
	return []map[string]interface{}{hash}
}

// Delete deletes a row or multiple rows using the contraint in extra
func (self *Model) Delete(extra ...url.Values) error {
	if err := self.deleteHash(extra...); err != nil {
		return err
	}

	self.aLISTS = []map[string]interface{}{make(map[string]interface{})}
	for k, v := range extra[0] {
		self.aLISTS[0][k] = v[0]
	}

	return nil
}

// properValue returns the value of key 'v' from extra.
// In case it does not exist, it tries to get from ARGS.
func (self *Model) properValue(v string, extra url.Values) interface{} {
	ARGS := self.aARGS
	if !hasValue(extra) {
		return ARGS.Get(v)
	}
	if val := extra.Get(v); val != "" {
		return val
	}
	return ARGS.Get(v)
}

// properValues returns the values of multiple keys 'vs' from extra.
// In case it does not exists, it tries to get from ARGS.
func (self *Model) properValues(vs []string, extra url.Values) []interface{} {
	ARGS := self.aARGS
	outs := make([]interface{}, len(vs))
	if !hasValue(extra) {
		for i, v := range vs {
			outs[i] = ARGS.Get(v)
		}
		return outs
	}
	for i, v := range vs {
		val := extra.Get(v)
		if val != "" {
			outs[i] = val
		} else {
			outs[i] = ARGS.Get(v)
		}
	}
	return outs
}

// properValuesHash is the same as properValues, but resulting in a map.
func (self *Model) properValuesHash(vs []string, extra url.Values) map[string]interface{} {
	values := self.properValues(vs, extra)
	hash := make(map[string]interface{})
	for i, v := range vs {
		hash[v] = values[i]
	}
	return hash
}

// OrderString outputs the ORDER BY string using information in args
func (self *Model) OrderString() string {
	ARGS := self.aARGS
	column := ""
	if ARGS.Get(self.Sortby) != "" {
		column = ARGS.Get(self.Sortby)
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
	if ARGS.Get(self.Sortreverse) != "" {
		order += " DESC"
	}

	if ARGS.Get(self.Rowcount) != "" {
		rowcount, err := strconv.Atoi(ARGS.Get(self.Rowcount))
		if err != nil {
			return ""
		}
		pageno := 1
		if ARGS.Get(self.Pageno) == "" {
			ARGS.Set(self.Pageno, "1")
		} else {
			pageno, err = strconv.Atoi(ARGS.Get(self.Pageno))
			if err != nil {
				return ""
			}
		}
		order += " LIMIT " + ARGS.Get(self.Rowcount) + " OFFSET " + strconv.Itoa((pageno-1)*rowcount)
	}

	matched, err := regexp.MatchString("[;'\"]", order)
	if err != nil || matched {
		return ""
	}
	return order
}
