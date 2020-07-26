package godbi

import (
	"net/url"
	"errors"
)

type Schema struct {
	Models	map[string]*Model
	Actions map[string]map[string]interface{}
}

// CallOnce calls page's action once and place data as a marker in OTHER
func (self *Model) CallOnce(item map[string]interface{}, page *Page, extra ...url.Values) error {
	modelName  := page.Model
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

	schema        := self.Schema
	model_obj, ok := schema.Models[modelName]
	if !ok {
		return errors.New("1081")
    }
	action_funcs, ok := schema.Actions[modelName]
	if !ok {
		return errors.New("1082")
	}
	action_func, ok := action_funcs[actionName]
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
			extra = make([]url.Values,1)
		}
		extra[0] = hash
	}

	model_obj.UpdateModel(self.Db, args, schema)
	final_action := action_func.(func(...url.Values) error)
	if err := final_action(extra...); err != nil {
		return err
	}

	lists := model_obj.LISTS
	if HasValue(lists) {
		item[marker] = lists
		model_obj.LISTS = nil
	}

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
			extra = make([]url.Values,1)
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

	nextpages, ok := self.Nextpages[action];
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
