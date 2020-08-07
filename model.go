package godbi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math"
	"math/rand"
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
	Crud
	Navigate

	// Actions: map between name and action functions
	Actions map[string]func(...url.Values) error

	// aARGS: the input data received by the web request
	aARGS url.Values

	// aLISTS: output data as slice of map, which represents a table row
	aLISTS []map[string]interface{}

	// Nextpages: defining how to call other models' actions
	Nextpages      map[string][]*Page `json:"nextpages,omitempty"`
	CurrentIDAuto  string             `json:"current_id_auto,omitempty"`
	InsertPars     []string           `json:"insert_pars,omitempty"`
	EditPars       []string           `json:"edit_pars,omitempty"`
	UpdatePars     []string           `json:"update_pars,omitempty"`
	InsupdPars     []string           `json:"insupd_pars,omitempty"`
	TopicsPars     []string           `json:"topics_pars,omitempty"`
	TopicsHashPars map[string]string  `json:"topics_hash,omitempty"`

	TotalForce  int    `json:"total_force,omitempty"`
	Empties     string `json:"empties,omitempty"`
	Fields      string `json:"fields,omitempty"`
	Maxpageno   string `json:"maxpageno,omitempty"`
	Totalno     string `json:"totalno,omitempty"`
	Rowcount    string `json:"rawcount,omitempty"`
	Pageno      string `json:"pageno,omitempty"`
	Sortreverse string `json:"sortreverse,omitempty"`
	Sortby      string `json:"sortby,omitempty"`
}

// NewModel creates a new Model struct from json file 'filename'
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
			return self.ProperValues(self.CurrentKeys, extra[0])
		}
		return self.ProperValues(self.CurrentKeys, nil)
	}
	if hasValue(extra) {
		return []interface{}{self.ProperValue(self.CurrentKey, extra[0])}
	}
	return []interface{}{self.ProperValue(self.CurrentKey, nil)}
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
			if err := self.TotalHash(&nt, extra...); err != nil {
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
	if self.TopicsHashPars == nil {
		fields = self.filteredFields(self.TopicsPars)
	} else {
		fields = self.TopicsHashPars
	}
	self.aLISTS = make([]map[string]interface{}, 0)
	return self.TopicsHashOrder(&self.aLISTS, fields, self.OrderString(), extra...)
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
		return self.EditHash(&self.aLISTS, fields, val, extra[0])
	}
	return self.EditHash(&self.aLISTS, fields, val)
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
	if err := self.InsertHash(fieldValues); err != nil {
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

	if err := self.InsupdTable(fieldValues, uniques); err != nil {
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

	if err := self.UpdateHashNulls(fieldValues, val, self.aARGS[self.Empties], extra...); err != nil {
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
	if err := self.DeleteHash(extra...); err != nil {
		return err
	}

	self.aLISTS = []map[string]interface{}{make(map[string]interface{})}
	for k, v := range extra[0] {
		self.aLISTS[0][k] = v[0]
	}

	return nil
}

// Existing checks if table has val in field
func (self *Model) Existing(table string, field string, val interface{}) error {
	id := 0
	return self.DB.QueryRow("SELECT "+field+" FROM "+table+" WHERE "+field+"=?", val).Scan(&id)
}

// Randomid create PK field's int value that does not exists in the table
func (self *Model) Randomid(table string, field string, m ...interface{}) (int, error) {
	ARGS := self.aARGS
	var min, max, trials int
	if m == nil {
		min = 0
		max = 4294967295
		trials = 10
	} else {
		min = m[0].(int)
		max = m[1].(int)
		if m[2] == nil {
			trials = 10
		} else {
			trials = m[2].(int)
		}
	}

	for i := 0; i < trials; i++ {
		val := min + int(rand.Float32()*float32(max-min))
		if err := self.Existing(table, field, val); err != nil {
			continue
		}
		ARGS.Set(field, strconv.FormatInt(int64(val), 10))
		return val, nil
	}

	return 0, errors.New("cannot get a random id")
}

// ProperValue returns the value of key 'v' from extra.
// In case it does not exist, it tries to get from ARGS.
func (self *Model) ProperValue(v string, extra url.Values) interface{} {
	ARGS := self.aARGS
	if !hasValue(extra) {
		return ARGS.Get(v)
	}
	if val := extra.Get(v); val != "" {
		return val
	}
	return ARGS.Get(v)
}

// ProperValues returns the values of multiple keys 'vs' from extra.
// In case it does not exists, it tries to get from ARGS.
func (self *Model) ProperValues(vs []string, extra url.Values) []interface{} {
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

// ProperValuesHash is the same as ProperValues, but resulting in a map.
func (self *Model) ProperValuesHash(vs []string, extra url.Values) map[string]interface{} {
	values := self.ProperValues(vs, extra)
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
