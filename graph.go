package godbi

import (
	"context"
	"database/sql"
	"fmt"
)

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
// 'model' is the model name, and 'action' the action name.
// The first extra is the input data, shared by all sub actions.
// The rest are specific data for each action starting with the current one.
//
func (self *Graph) Run(model, action string, ARGS map[string]interface{}, extra ...interface{}) ([]map[string]interface{}, error) {
	return self.RunContext(context.Background(), model, action, ARGS, extra...)
}

// RunContext runs action by model and action string names.
// It returns the searched data and optional error code.
//
// 'model' is the model name, and 'action' the action name.
// The first extra is the input data, shared by all sub actions.
// The rest are specific data for each action starting with the current one.
//
func (self *Graph) RunContext(ctx context.Context, model, action string, ARGS map[string]interface{}, extra ...interface{}) ([]map[string]interface{}, error) {
	modelObj, ok := self.Models[model]
	if !ok {
		return nil, fmt.Errorf("model %s not found in graph", model)
	}

	nones := modelObj.NonePass(action)
	// nones input should be moved from extra to ARGS
	if nones != nil && hasValue(extra) && hasValue(extra[0]) {
		switch v := extra[0].(type) {
		case map[string]interface{}:
			newExtra := make(map[string]interface{})
			for key, value := range v {
				for _, item := range nones {
					if item == key {
						ARGS[key] = value
					} else {
						newExtra[key] = value
					}
				}
			}
			extra[0] = newExtra
		default:
		}
	}

	lists, nextpages, err := modelObj.RunModelContext(ctx, self.DB, action, ARGS, extra...)
	if err != nil {
		return nil, err
	}

	// nones input should not be passed to the next page
	// in the next page, these parameters are assigned from extra
	if nones != nil {
		for _, item := range nones {
			if _, ok := ARGS[item]; ok {
				delete(ARGS, item)
			}
		}
	}

	if !hasValue(lists) || nextpages == nil {
		return lists, nil
	}

	for _, page := range nextpages {
		if hasValue(extra) {
			extra = extra[1:]
		}
		//extra0 := make(map[string]interface{})
		var extra0 interface{}
		if hasValue(extra) {
			extra0 = extra[0]
		}
		extra0 = page.manualRefresh(extra0)
		for _, item := range lists {
			newExtra0, ok := page.refresh(item, extra0)
			if !ok {
				continue
			}
			newExtras := []interface{}{newExtra0}
			if hasValue(extra) {
				newExtras = append(newExtras, extra[:1]...)
			}
			newLists, err := self.RunContext(ctx, page.TableName, page.ActionName, ARGS, newExtras...)
			if err != nil {
				return nil, err
			}
			item[page.TableName+"_"+page.ActionName] = newLists
		}
	}

	return lists, nil
}
