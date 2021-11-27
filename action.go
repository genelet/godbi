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
	GetName() string
	GetAppendix() interface{}
	SetMusts([]string)
	SetNextpage([]*Nextpage)
	RunActionContext(context.Context, *sql.DB, *Table, map[string]interface{}, ...interface{}) ([]map[string]interface{}, []*Nextpage, error)
}

type Action struct {
	ActionName string     `json:"actionName,omitempty" hcl:"actionName,optional"`
	Musts     []string    `json:"musts,omitempty" hcl:"musts,optional"`
	Nextpages []*Nextpage `json:"nextpages,omitempty" hcl:"nextpages,block"`
	Appendix  interface{} `json:"appendix,omitempty" hcl:"appendix,block"`
}

func (self *Action) SetMusts(musts []string) {
	self.Musts = musts
}

func (self *Action) SetNextpage(edges []*Nextpage) {
	self.Nextpages = edges
}

func (self *Action) GetName() string {
	return self.ActionName
}

func (self *Action) GetAppendix() interface{} {
	return self.Appendix
}

func (self *Action) filterPars(currentTable string, ARGS map[string]interface{}, rename []*Col, fieldsName string, joins []*Joint) (string, []interface{}, string) {
	var fields []string
	if v, ok := ARGS[fieldsName]; ok {
		fields = v.([]string)
	}

	keys := make([]string, 0)
	labels := make([]interface{}, 0)
	for _, col := range rename {
		if fields==nil || grep (fields, col.ColumnName) {
			keys = append(keys, col.ColumnName)
			labels = append(labels, [2]string{col.Label, col.TypeName})
		}
	}
	sql := strings.Join(keys, ", ")

	var table string
	if hasValue(joins) {
		sql = "SELECT " + sql + "\nFROM " + joinString(joins)
		table = joins[0].getAlias()
	} else {
		sql = "SELECT " + sql + "\nFROM " + currentTable
	}

	return sql, labels, table
}

func (self *Action) checkNull(ARGS map[string]interface{}, extra ...interface{}) error {
	if self.Musts == nil {
		return nil
	}
	for _, item := range self.Musts {
		err := fmt.Errorf("item %s not found in input", item)
		if _, ok := ARGS[item]; !ok {
			if hasValue(extra) && hasValue(extra[0]) {
				switch t := extra[0].(type) {
				case []map[string]interface{}:
					for _, each := range t {
						if _, ok = each[item]; !ok {
							return err
						}
					}
				case map[string]interface{}:
					if _, ok = t[item]; !ok {
						return err
					}
				default:
					return err
				}
			} else {
				return err
			}
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
