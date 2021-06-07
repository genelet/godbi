package godbi

import (
	"database/sql"
	"errors"
)

// Schema describes all models and actions in a database schema
//
type Schema struct {
	db     *sql.DB
	Models map[string]Navigate
}

func NewSchema(db *sql.DB, s map[string]Navigate) *Schema {
	return &Schema{db, s}
}

// Run runs action by model and action string names.
// It returns the searched data and optional error code.
//
// Model is the model name, and action the action name.
// The first extra is the input data, shared by all sub actions.
// The rest are specific data for each actions starting with the current one.
//
func (self *Schema) Run(model, action string, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
	modelObj, ok := self.Models[model]
	if !ok {
		return nil, errors.New("model not found in schema models")
	}

	var args map[string]interface{}
	if hasValue(extra) {
		args = extra[0]
		extra = extra[1:] // shift immediately to make sure ARGS not in extra
		nones := modelObj.nonePass() // move none passed pars to ARGS
		if hasValue(extra) && hasValue(extra[0]) {
			for _, item := range nones {
				if fs, ok := extra[0][item]; ok {
					args[item] = fs
					delete(extra[0], item)
				}
			}
		}
	}
	modelObj.SetDB(self.db)
	modelObj.SetArgs(args)
	act := modelObj.GetAction(action)
	if act == nil {
		return nil, errors.New("action not found in schema model")
	}

	if err := act(extra...); err != nil {
		return nil, err
	}
	lists := modelObj.CopyLists()
	modelArgs := modelObj.getArgs(true) // for nextpages to use
	nextpages := modelObj.getNextpages(action)

	modelObj.SetArgs(nil)
	modelObj.SetDB(nil)
	modelObj.cleanLists()


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
