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
			if item.GetTable().GetTableName() == model {
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
			switch t := rest[1].(type) {
			case map[string]interface{}: extra = t
			default:
				return nil, fmt.Errorf("Wrong type for data: %#v", rest[0])
			}
		}
	}


	if self.argsMap[model] != nil {
		argsMap := self.argsMap[model].(map[string]interface{})
		args = MergeArgs(args, argsMap[action])
	}

	if self.extraMap[model] != nil {
		extraAction := self.extraMap[model].(map[string]interface{})
		if extraAction[action] != nil {
			extra = MergeExtra(extra, extraAction[action].(map[string]interface{}))
		}
	}

	switch t := args.(type) {
	case map[string]interface{}:
		return self.hashContext(ctx, db, model, action, t, extra)
	case []map[string]interface{}:
		var final []map[string]interface{}
		for _, arg := range t {
			lists, err := self.hashContext(ctx, db, model, action, arg, extra)
			if err != nil { return nil, err }
			final = append(final, lists...)
		}
		return final, nil
	case []interface{}:
		var final []map[string]interface{}
		for _, arg := range t {
			if v, ok := arg.(map[string]interface{}); ok {
				lists, err := self.hashContext(ctx, db, model, action, v, extra)
				if err != nil { return nil, err }
				final = append(final, lists...)
			}
		}
		return final, nil
	default:
	}

	return self.hashContext(ctx, db, model, action, nil, extra)
}

// RunContext runs action by model and action string names.
// It returns the searched data and optional error code.
//
// 'model' is the model name, and 'action' the action name.
// The first extra is the input data, shared by all sub actions.
// The rest are specific data for each action starting with the current one.
//
func (self *Graph) hashContext(ctx context.Context, db *sql.DB, model, action string, args, extra map[string]interface{}) ([]map[string]interface{}, error) {
	modelObj := self.GetModel(model)
	if modelObj == nil {
		return nil, fmt.Errorf("model %s not found in graph", model)
	}

	actionObj := modelObj.GetAction(action)
	if actionObj == nil {
		return nil, fmt.Errorf("action %s not found in graph", action)
	}

	if args != nil && actionObj.GetIsDo() {
		args = modelObj.GetTable().RefreshArgs(args).(map[string]interface{})
	}

	prepares := actionObj.GetPrepares()
	nextpages := actionObj.GetNextpages()

	newArgs := CloneArgs(args)
	newExtra := CloneExtra(extra)
	// prepares receives filtered args and extra from current args
	if prepares != nil {
		for _, p := range prepares {
			// in case of prepare, we use args to get
			// NextArgs and NextExtra as nextpage's input and constrains
			preArgs := CloneArgs(args)
			preExtra := CloneExtra(extra)
			if p.TableName != model {
				v, ok := p.FindArgs(preArgs)
				pAction := self.GetModel(p.TableName).GetAction(p.ActionName)
				if pAction.GetIsDo() && ok && !hasValue(v) {
					continue
				}
				preArgs = MergeArgs(p.NextArgs(preArgs), v)
				preExtra = MergeExtra(p.NextExtra(preArgs), p.FindExtra(preExtra))
			}
			lists, err := self.RunContext(ctx, db, p.TableName, p.ActionName, preArgs, preExtra)
			if err != nil { return nil, err }
			// only two types of prepares
			// 1) one pre, with multiple outputs (when p.argsMap is multiple)
			if hasValue(lists) && len(lists) > 1 {
				var tmp []map[string]interface{}
				newExtra = CloneExtra(extra)
				for _, item := range lists {
					result := MergeArgs(args, p.NextArgs(item)).(map[string]interface{})
					tmp = append(tmp, result)
					newExtra = MergeExtra(newExtra, p.NextExtra(item))
				}
				newArgs = tmp
				break
			}
			// 2) multiple pre, with one output each.
			// when a multiple output is found, 1) will override
			if hasValue(lists) && hasValue(lists[0]) {
				newArgs = MergeArgs(newArgs, p.NextArgs(lists[0]).(map[string]interface{}))
				newExtra = MergeExtra(newExtra, p.NextExtra(lists[0]))
			}
		}
	}

	data, err := modelObj.RunModelContext(ctx, db, action, newArgs, newExtra)
	if err != nil { return nil, err }

	if nextpages == nil {
		return data, nil
	}

	for _, p := range nextpages {
		for _, item := range data {
			v, ok := p.FindArgs(newArgs)
			pAction := self.GetModel(p.TableName).GetAction(p.ActionName)
			// is a do-action, needs input from the table, but not found
			if pAction.GetIsDo() && ok && !hasValue(v) {
				continue
			}
			nextArgs  := MergeArgs(p.NextArgs(item), v)
			nextExtra := MergeExtra(p.NextExtra(item), p.FindExtra(newExtra))
			newLists, err := self.RunContext(ctx, db, p.TableName, p.ActionName, nextArgs, nextExtra)
			if err != nil { return nil, err }
			if hasValue(newLists) {
				item[p.Subname()] = newLists
			}
		}
	}

	return data, nil
}
