package godbi

import (
"log"
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
			if item.GetTableName() == model {
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
log.Printf("1 model %s action %s rest %#v", model, action, rest)
	var args interface{}
	var extra map[string]interface{}
	if rest != nil {
		args = rest[0]
		if len(rest) == 2 {
			switch t := rest[1].(type) {
			case map[string]interface{}: extra = t
			default:
				return nil, fmt.Errorf("Wrong type for data: %#v", rest[0])
			}
		}
	}

log.Printf("2 args: %#v, extra: %#v", args, extra)
	modelObj := self.GetModel(model)
	if modelObj == nil {
		return nil, fmt.Errorf("model %s not found in graph", model)
	}

log.Printf("%#v", 3)
	actionObj := modelObj.GetAction(action)
	if actionObj == nil {
		return nil, fmt.Errorf("action %s not found in graph", action)
	}
	prepares := actionObj.GetPrepares()
	nextpages := actionObj.GetNextpages()

	newArgs := CloneArgs(args)
	newExtra := CloneExtra(extra)

log.Printf("%#v", 4)
	// prepares only affect args, and no RelateArgs nor RelateExtra involed
	if prepares != nil {
		for _, p := range prepares {
			lists, err := self.RunContext(ctx, db, p.TableName, p.ActionName, rest...)
			if err != nil { return nil, err }
			// only two types of prepares
			// 1) one pre, with multiple outputs (when p.argsMap is multiple)
			if hasValue(lists) && len(lists) > 0 {
				newArgs = CloneArgs(args)
				newExtra = CloneExtra(extra)
				for _, item := range lists {
					newArgs = p.MergeArgs(newArgs, item)
					newExtra = p.MergeExtra(newExtra, item)
				}
				break
			}
			// 2) multiple pre, with one output each.
			// when a multiple output is found, 1) will override
			newArgs = p.MergeArgs(newArgs, lists[0])
			newExtra = p.MergeExtra(newExtra, lists[0])
		}
	}

log.Printf("%#v", 5)
	var data []map[string]interface{}
	var err error

	if self.extraMap[model] != nil {
		extraAction := self.extraMap[model].(map[string]interface{})
		if extraAction[action] != nil {
			newExtra = MergeExtra(newExtra, extraAction[action].(map[string]interface{}))
		}
	}

log.Printf("%#v", 6)
	var argsModelAction interface{}
	if self.argsMap[model] != nil {
		argsMap := self.argsMap[model].(map[string]interface{})
		argsModelAction = argsMap[action]
	}
	if argsModelAction != nil {
log.Printf("%#v", 7)
		switch t := argsModelAction.(type) {
		case map[string]interface{}:
			data, err = modelObj.RunModelContext(ctx, db, action, MergeArgs(newArgs, t), newExtra)
			if err != nil { return nil, err }
		case []map[string]interface{}:
			for _, each := range t {
				lists, err := modelObj.RunModelContext(ctx, db, action, MergeArgs(newArgs, each), newExtra)
				if err != nil { return nil, err }
				data = append(data, lists...)
			}
		default:
			return nil, fmt.Errorf("wrong input data type: %#v", t)
		}
	} else {
log.Printf("%#v", 8)
		data, err = modelObj.RunModelContext(ctx, db, action, newArgs, newExtra)
		if err != nil { return nil, err }
	}

log.Printf("9 data: %#v", data)
	if nextpages == nil { return data, nil }

log.Printf("10 next page: %#v", nextpages)
	for _, p := range nextpages {
log.Printf("101 next page: %#v", p)
		for _, item := range data {
log.Printf("11 %#v", item)
log.Printf("12 %#v", p)
			newLists, err := self.RunContext(ctx, db, p.TableName, p.ActionName, p.NextArgs(item), p.NextExtra(item))
			if err != nil { return nil, err }
log.Printf("13 sub list: %#v", newLists)
			item[p.Subname()] = newLists
		}
	}

log.Printf("14")
	return data, nil
}
