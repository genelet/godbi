package godbi

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// Action is to implement Capability interface
//
type Capability interface {
	RunActionContext(context.Context, *sql.DB, *Table, map[string]interface{}, ...interface{}) ([]map[string]interface{}, []*Edge, error)
}

type Action struct {
	Must      []string    `json:"must,omitempty" hcl:"must,optional"`
	Nextpages []*Edge     `json:"nextpages,omitempty" hcl:"nextpage,block"`
	Appendix  interface{} `json:"appendix,omitempty" hcl:"appendix,block"`
}

func (self *Action) filterPars(currentTable string, ARGS map[string]interface{}, rename map[string][]string, fieldsName string, joins []*Join) (string, []interface{}, string) {
	var fields []string
	if v, ok := ARGS[fieldsName]; ok {
		fields = v.([]string)
	}

	shorts := make(map[string][2]string)
	for k, v := range rename {
		if fields == nil {
			shorts[k] = [2]string{v[0], v[1]}
		} else {
			if grep(fields, k) {
				shorts[k] = [2]string{v[0], v[1]}
			}
		}
	}

	sql, labels := selectType(shorts)
	var table string
	if hasValue(joins) {
		sql = "SELECT " + sql + "\nFROM " + joinString(joins)
		table = joins[0].getAlias()
	} else {
		sql = "SELECT " + sql + "\nFROM " + currentTable
	}

	return sql, labels, table
}

func (self *Action) checkNull(ARGS map[string]interface{}) error {
	if self.Must == nil {
		return nil
	}
	for _, item := range self.Must {
		if _, ok := ARGS[item]; !ok {
			return fmt.Errorf("item %s not found in input", item)
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
	if fieldNames == nil {
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

	keys := make([]string, 0)
	labels := make([]interface{}, 0)
	for key, val := range selectPars {
		keys = append(keys, key)
		labels = append(labels, val)
	}
	return strings.Join(keys, ", "), labels
}
