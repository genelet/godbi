package godbi

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"fmt"
)

// Graph describes all models and actions in a database schema
//
type Graph struct {
	Models []Navigate `json:"models" hcl:"models"`
	argsMap map[string]interface{}
	extraMap map[string]interface{}
}

func NewGraphJsonFile(fn string, cmap ...map[string][]Capability) (*Graph, error) {
	dat, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	return NewGraphJson(json.RawMessage(dat), cmap...)
}

type g struct {
	Models []*m `json:"models" hcl:"models"`
}

func NewGraphJson(dat json.RawMessage, cmap ...map[string][]Capability) (*Graph, error) {
	tmps := new(g)
	if err := json.Unmarshal(dat, &tmps); err != nil {
		return nil, err
	}

	var models []Navigate
	for _, tmp := range tmps.Models {
		var cs []Capability
		if cmap != nil && cmap[0] != nil {
			cs, _ = cmap[0][tmp.TableName]
		}
		actions, err := Assertion(tmp.Actions, cs...)
		if err != nil {
			return nil, err
		}
		models = append(models, &Model{tmp.Table, actions})
	}

	return &Graph{Models:models}, nil
}

func (self *Graph) Initialize(args map[string]interface{}, extra map[string]interface{}) {
	self.argsMap = args
    self.extraMap = extra
}

func (self *Graph) GetModel(model string) Navigate {
	if self.Models != nil {
		for _, item := range self.Models {
			if item.GetName() == model {
				return item
			}
		}
	}
	return nil
}

// RunContext runs action by model and action string names.
// It returns the searched data and optional error code.
//
// 'model' is the model name, and 'action' the action name.
// The first extra is the input data, shared by all sub actions.
// The rest are specific data for each action starting with the current one.
//
func (self *Graph) RunContext(ctx context.Context, db *sql.DB, model, action string, rest ...interface{}) ([]map[string]interface{}, error) {
	var args interface{}
	var extra map[string]interface{}
	if rest != nil {
		args = rest[0]
		if len(rest) == 2 {
			switch t := rest[0].(type) {
			case map[string]interface{}: extra = t
			default:
				return nil, fmt.Errorf("Wrong type for data: %#v", rest[0])
			}
		}
	}

	modelObj := self.GetModel(model)
	if modelObj == nil {
		return nil, fmt.Errorf("model %s not found in graph", model)
	}

	actionObj := modelObj.GetAction(action)
	if actionObj == nil {
		return nil, fmt.Errorf("action %s not found in graph", action)
	}
	prepares := actionObj.GetPrepares()
	nextpages := actionObj.GetNextpages()

	newArgs := cloneArgs(args)

	// prepares only affect args, and no RelateArgs nor RelateExtra involed
	if prepares != nil {
		for _, p := range prepares {
			lists, err := self.RunContext(ctx, db, p.TableName, p.ActionName)
			if err != nil { return nil, err }
			// only two types of prepares
			// 1) one pre, with multiple outputs (when p.argsMap is multiple)
			if hasValue(lists) && len(lists) > 0 {
				newArgs = cloneArgs(args)
				for _, item := range lists {
					newArgs = appendArgs(newArgs, item)
				}
				break
			}
			// 2) multiple pre, with one output each.
			// when a multiple output is found, 1) will override
			newArgs = appendArgs(newArgs, lists[0])
		}
	}

	newExtra := cloneMap(extra)
	if self.extraMap[model] != nil {
		extraAction := self.extraMap[model].(map[string]interface{})
		if extraAction[action] != nil {
			newExtra = appendMap(newExtra, extraAction[action].(map[string]interface{}))
		}
	}

	var argsModelAction interface{}
	if self.argsMap[model] != nil {
		argsMap := self.argsMap[model].(map[string]interface{})
		argsModelAction = argsMap[action]
	}

	var data []map[string]interface{}
	var err error
	if argsModelAction != nil {
		switch t := argsModelAction.(type) {
		case map[string]interface{}:
			data, err = modelObj.RunModelContext(ctx, db, action, appendArgs(newArgs, t), newExtra)
			if err != nil { return nil, err }
		case []map[string]interface{}:
			for _, each := range t {
				lists, err := modelObj.RunModelContext(ctx, db, action, appendArgs(newArgs, each), newExtra)
				if err != nil { return nil, err }
				data = append(data, lists...)
			}
		default:
			return nil, fmt.Errorf("original input wrong %#v", t)
		}
	} else {
		data, err = modelObj.RunModelContext(ctx, db, action, newArgs, newExtra)
		if err != nil { return nil, err }
	}

	if nextpages == nil { return data, nil }

	for _, p := range nextpages {
		for _, item := range data {
			newLists, err := self.RunContext(ctx, db, p.TableName, p.ActionName, p.NextArgs(item), p.NextExtra(item))
			if err != nil { return nil, err }
			item[p.Subname()] = newLists
		}
	}

	return data, nil
}
