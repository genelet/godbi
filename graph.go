package godbi

import (
	"database/sql"
	"errors"
)

// Navigate is interface to implement Model
//
type Navigate interface {
	// NonePass: keys in restriction extra which should not be passed to the nextpage but assign to this args only
	NonePass() []string

	// SetArgs: set new input
	SetArgs(map[string]interface{})

	// SetDB: set SQL handle
	SetDB(*sql.DB)

	// RunAction runs action by name with optional restrictions
	RunAction(string, ...map[string]interface{}) ([]map[string]interface{}, map[string]interface{}, []*Page, error)
}

// Graph describes all models and actions in a database schema
//
type Graph struct {
	*sql.DB
	Models map[string]Navigate
}

func NewGraph(db *sql.DB, s map[string]Navigate) *Graph {
	return &Graph{db, s}
}

// Run runs action by model and action string names.
// It returns the searched data and optional error code.
//
// Model is the model name, and action the action name.
// The first extra is the input data, shared by all sub actions.
// The rest are specific data for each actions starting with the current one.
//
func (self *Graph) Run(model, action string, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
	modelObj, ok := self.Models[model]
	if !ok {
		return nil, errors.New("model not found in schema models")
	}

	var args map[string]interface{}
	if hasValue(extra) {
		args = extra[0]
		extra = extra[1:] // shift immediately to make sure ARGS not in extra
		nones := modelObj.NonePass() // move none passed pars to ARGS
		if hasValue(extra) && hasValue(extra[0]) {
			for _, item := range nones {
				if fs, ok := extra[0][item]; ok {
					args[item] = fs
					delete(extra[0], item)
				}
			}
		}
	}

	modelObj.SetDB(self.DB)
	modelObj.SetArgs(args)
	lists, modelArgs, nextpages, err := modelObj.RunAction(action, extra...)
	modelObj.SetArgs(nil)
	modelObj.SetDB(nil)
	if err != nil {
		return nil, err
	}

	if !hasValue(lists) || nextpages == nil {
		return lists, nil
	}

	for _, page := range nextpages {
		if hasValue(extra) {
			extra = extra[1:]
		}
		extra0 := make(map[string]interface{})
		if hasValue(extra) {
			extra0 = extra[0]
		}
		if page.Extra != nil { // use name Extra instead of Manual
			for k, v := range page.Extra {
				extra0[k] = v
			}
		}
		for _, item := range lists {
			newExtra0, ok := page.refresh(item, extra0)
			if !ok {
				continue
			}
			//newExtras := []map[string]interface{}{newExtra0}
			newExtras := []map[string]interface{}{modelArgs, newExtra0}
			if hasValue(extra) {
				newExtras = append(newExtras, extra[:1]...)
			}
			//newLists, err := self.Run(page.Model, page.Action, modelArgs, newExtras...)
			newLists, err := self.Run(page.Model, page.Action, newExtras...)
			if err != nil {
				return nil, err
			}
			item[page.Model+"_"+page.Action] = newLists
		}
	}

	return lists, nil
}
