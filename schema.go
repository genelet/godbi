package godbi

import (
	"database/sql"
	"errors"
	"net/url"
)

// Restful is interface whose methods have been implemented in Model
// GetLists: get the main data as slice of rows which is a map in column name and column value
// UpdateModel: pass db handle, args and schema into Model and set the data to be 0-sized
type Restful interface {
    GetLists() []map[string]interface{}
	UpdateModel(*sql.DB, url.Values, *Schema)
}

// Schema describes all models and actions in a database schema
// Models: map between model name and model struct
// Actions: map between model name and actions,
// represented as a map between action name and action method
type Schema struct {
	Models  map[string]Restful
	Actions map[string]map[string]interface{}
}

// Page type describes next page's structure
// Model: the name of the model
// Action: the method name on the model
// Alias: the retrieved data is assigned to key: model_action as default. Use Alias to replace it.
// Ignore: if the key exists, don't run the next page
// Manual: constraint conditions manually assigned
// RelateItem: current page's column versus next page's column. The value is forced as constraint.
type Page struct {
	Model	string                 `json:"model"`
	Action	string                 `json:"action"`
	Alias	string                 `json:"alias,omitempty"`
	Ignore	bool                   `json:"ignore,omitempty"`
	Manual	map[string]string      `json:"manual,omitempty"`
	RelateItem map[string]string   `json:"relate_item,omitempty"`
}

// CallOnce calls page's action once and places data as a marker in item
func (self *Model) CallOnce(item map[string]interface{}, page *Page, extra ...url.Values) error {
	modelName := page.Model
	actionName := page.Action

	marker := modelName + "_" + actionName
	if page.Alias != "" {
		marker = page.Alias
	}
	if page.Ignore {
		if _, ok := item[marker]; ok {
			return nil
		}
	}

	schema := self.Scheme
	modelObj, ok := schema.Models[modelName]
	if !ok {
		return errors.New("1081")
	}
	actionFuncs, ok := schema.Actions[modelName]
	if !ok {
		return errors.New("1082")
	}
	actionFunc, ok := actionFuncs[actionName]
	if !ok {
		return errors.New("1083")
	}

	args := url.Values{}
	for k, v := range self.ARGS {
		if Grep([]string{self.Sortby, self.Sortreverse, self.Rowcount, self.Totalno, self.Pageno, self.Maxpageno}, k) {
			continue
		}
		args[k] = v
	}

	hash := url.Values{}
	if HasValue(extra) {
		for k, v := range extra[0] {
			hash[k] = v
		}
	}
	if page.Manual != nil {
		for k, v := range page.Manual {
			hash.Set(k, v)
		}
	}
	if HasValue(hash) {
		if !HasValue(extra) {
			extra = make([]url.Values, 1)
		}
		extra[0] = hash
	}

	modelObj.UpdateModel(self.Db, args, schema)
	finalAction := actionFunc.(func(...url.Values) error)
	if err := finalAction(extra...); err != nil {
		return err
	}

	lists := modelObj.GetLists()
	if HasValue(lists) {
		item[marker] = lists
	}
	modelObj.UpdateModel(nil, nil, nil)

	return nil
}

// CallNextpage calls page's action, for each item in LISTS.
func (self *Model) CallNextpage(page *Page, extra ...url.Values) error {
	lists := self.LISTS
	if !HasValue(lists) || !HasValue(page.RelateItem) {
		return nil
	}

	for _, item := range lists {
		if !HasValue(extra) {
			extra = make([]url.Values, 1)
			extra[0] = url.Values{}
		}
		found := false
		for k, v := range page.RelateItem {
			if t, ok := item[k]; ok {
				found = true
				extra[0].Set(v, Interface2String(t))
			}
		}
		if found == false {
			continue
		}
		if err := self.CallOnce(item, page, extra...); err != nil {
			return err
		}
	}

	return nil
}

// ProcessAfter calls all pages' actions, defined in Nextpages.
// each action's value is placed in LISTS as a key-value pair
func (self *Model) ProcessAfter(action string, extra ...url.Values) error {
	if !HasValue(self.Nextpages) {
		return nil
	}

	nextpages, ok := self.Nextpages[action]
	if !ok {
		return nil
	}

	for _, page := range nextpages {
		if HasValue(extra) {
			extra = extra[1:]
		}
		if err := self.CallNextpage(page, extra...); err != nil {
			return err
		}
	}
	return nil
}
