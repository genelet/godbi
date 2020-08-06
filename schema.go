package godbi

import (
	"database/sql"
	"errors"
	"net/url"
)

// Schema describes all models and actions in a database schema
//
type Schema struct {
	Models map[string]Navigate
}

// Page type describes next page's structure
// Model: the name of the model
// Action: the method name on the model
// Manual: constraint conditions manually assigned
// RelateItem: current page's column versus next page's column. The value is forced as constraint.
type Page struct {
	Model      string            `json:"model"`
	Action     string            `json:"action"`
	Manual     map[string]string `json:"manual,omitempty"`
	RelateItem map[string]string `json:"relate_item,omitempty"`
}

// Run runs action by model and action string names
// args: the input data
// db: the database handle
// extra: optional extra parameters
// The output are data and optional error code
//
func (self *Schema) Run(model, action string, args url.Values, db *sql.DB, extra ...url.Values) ([]map[string]interface{}, error) {
	modelObj, ok := self.Models[model]
	if !ok {
		return nil, errors.New("model not found in schema models")
	}

	modelObj.SetArgs(args)
	modelObj.SetDB(db)

	if err := modelObj.RunAction(action, extra...); err != nil {
		return nil, err
	}
	lists := modelObj.GetLists()
	modelArgs := modelObj.GetArgs(true) // for nextpages to use

	modelObj.SetArgs(url.Values{})
	modelObj.SetDB(nil)

	if !hasValue(lists) {
		return lists, nil
	}

	nextpages := modelObj.GetNextpages(action)
	if nextpages == nil {
		return lists, nil
	}

	for _, page := range nextpages {
		if hasValue(extra) {
			extra = extra[1:]
		}
		extra0 := url.Values{}
		if hasValue(extra) {
			extra0 = extra[0]
		}
		if page.Manual != nil {
			for k, v := range page.Manual {
				extra0.Set(k, v)
			}
		}
		for _, item := range lists {
			newExtra0, ok := page.refresh(item, extra0)
			if !ok {
				continue
			}
			newExtras := []url.Values{newExtra0}
			if hasValue(extra) {
				newExtras = append(newExtras, extra[:1]...)
			}
			newLists, err := self.Run(page.Model, page.Action, modelArgs, db, newExtras...)
			if err != nil {
				return nil, err
			}
			item[page.Model+"_"+page.Action] = newLists
		}
	}

	return lists, nil
}

func (self *Page) refresh(item map[string]interface{}, extra url.Values) (url.Values, bool) {
	newExtra := extra
	found := false
	for k, v := range self.RelateItem {
		if t, ok := item[k]; ok {
			found = true
			newExtra.Set(v, interface2String(t))
			break
		}
	}
	return newExtra, found
}
