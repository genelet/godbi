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

// Run runs action by model and action string names
// args: the input data
// db: the database handle
// extra: optional extra parameters
// The output are data and optional error code
//
func (self *Schema) Run(model, action string, args map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
	modelObj, ok := self.Models[model]
	if !ok {
		return nil, errors.New("model not found in schema models")
	}
	nones := modelObj.nonePass()
	if hasValue(extra) && hasValue(extra[0]) {
		for _, item := range nones {
			if fs, ok := extra[0][item]; ok {
				args[item] = fs
				delete(extra[0], item)
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
		if page.Manual != nil {
			for k, v := range page.Manual {
				extra0[k] = v
			}
		}
		for _, item := range lists {
			newExtra0, ok := page.refresh(item, extra0)
			if !ok {
				continue
			}
			newExtras := []map[string]interface{}{newExtra0}
			if hasValue(extra) {
				newExtras = append(newExtras, extra[:1]...)
			}
			newLists, err := self.Run(page.Model, page.Action, modelArgs, newExtras...)
			if err != nil {
				return nil, err
			}
			item[page.Model+"_"+page.Action] = newLists
		}
	}

	return lists, nil
}
