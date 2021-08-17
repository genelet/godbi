package godbi

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// Action is interface to implement Capability
//
type Action interface {
	Fulfill(string, []string, string, []string)
	RunContext(context.Context, *sql.DB, map[string]interface{}, ...map[string]interface{}) ([]map[string]interface{}, error)
	GetNextpages() []*Page
}

type Capability struct {
	Table
	Name       string   `json:"name" hcl:",label"`
	Must       []string `json:"must,omitempty" hcl:"must,optional"`
	Nextpages  []*Page  `json:"nextpages,omitempty" hcl:"nextpage,block"`
}

func (self *Capability) Fulfill(t string, pks []string, auto string, fks []string) {
	self.CurrentTable = t
	self.Pks          = pks
	self.IDAuto       = auto
	self.Fks          = fks
}

func (self *Capability) GetNextpages() []*Page {
	return self.Nextpages
}

func (self *Capability) CheckNull(ARGS map[string]interface{}) error {
	if self.Must == nil { return nil }
	for _, item := range self.Must {
		if _, ok := ARGS[item]; !ok {
			return fmt.Errorf("missing %s in input", item)
		}
	}
	return nil
}

func fromFv(fv map[string]interface{}) []map[string]interface{} {
    return []map[string]interface{}{fv}
}

// filteredFields outputs slice having elements in both inputs
// if the second fieldNames is null, output the first directly
//
func filteredFields(pars []string, fieldNames []string) []string {
	if fieldNames==nil {
		return pars
	}
    out := make([]string, 0)
    for _, field := range fieldNames {
        for _, v := range pars {
            if field == v {
                out = append(out, v)
                break
            }
        }
    }
    return out
}

func getFv(pars []string, ARGS map[string]interface{}, fieldNames []string) map[string]interface{} {
    fieldValues := make(map[string]interface{})
    for _, f := range filteredFields(pars, fieldNames) {
        if v, ok := ARGS[f]; ok {
            fieldValues[f] = v
        }
    }
    return fieldValues
}

func selectType(selectPars map[string][2]string) (string, []interface{}) {
	if selectPars == nil {
		return "", nil
	}

	keys   := make([]string, 0)
	labels := make([]interface{}, 0)
	for key, val := range selectPars {
		keys = append(keys, key)
		labels = append(labels, val)
	}
	return strings.Join(keys, ", "), labels
}
