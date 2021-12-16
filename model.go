package godbi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Model is to implement Navigate interface
//
type Navigate interface {
	GetTable() *Table
	GetAction(string) Capability
	RunModelContext(context.Context, *sql.DB, string, interface{}, ...map[string]interface{}) ([]map[string]interface{}, error)
}

type Model struct {
	Table
	Actions []Capability `json:"actions,omitempty" hcl:"actions,optional"`
}

func NewModelJsonFile(fn string, custom ...Capability) (*Model, error) {
	dat, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	return NewModelJson(json.RawMessage(dat), custom...)
}

type m struct {
	Table
	Actions []interface{} `json:"actions,omitempty"`
}
func NewModelJson(dat json.RawMessage, custom ...Capability) (*Model, error) {
	tmp := &m{}
	if err := json.Unmarshal(dat, tmp); err != nil {
		return nil, err
	}
	actions, err := Assertion(tmp.Actions, custom...)
	return &Model{tmp.Table, actions}, err
}

func Assertion(actions []interface{}, custom ...Capability) ([]Capability, error) {
	var trans []Capability

	for _, item := range actions {
		action := item.(map[string]interface{})
		v, ok := action["actionName"]
		if !ok { continue }
		name := ""
		switch u := v.(type) {
		case string: name = u
		default: return nil, fmt.Errorf("action name %v is wrongly typed", v)
		}

		jsonString, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}
		var tran Capability
		found := false
		for _, item := range custom {
			if name==item.GetActionName() {
				tran = item
				found = true
				break
			}
		}
		if !found {
			switch name {
			case "insert":
				tran = new(Insert)
			case "update":
				tran = new(Update)
			case "insupd":
				tran = new(Insupd)
			case "edit":
				tran = new(Edit)
			case "topics":
				tran = new(Topics)
			case "delete":
				tran = new(Delete)
			case "delecs":
				tran = new(Delecs)
			default:
				return nil, fmt.Errorf("action %s not defined", name)
			}
		}
		if err := json.Unmarshal(jsonString, tran); err != nil {
			return nil, err
		}
		trans = append(trans, tran)
	}
	return trans, nil
}

func (self *Model) GetTable() *Table {
	return &self.Table
}

func (self *Model) GetAction(action string) Capability {
	if self.Actions != nil {
		for _, item := range self.Actions {
			if item.GetActionName() == action {
				return item
			}
		}
	}

	return nil
}

func (self *Model) RunModel(db *sql.DB, action string, ARGS interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
	return self.RunModelContext(context.Background(), db, action, ARGS, extra...)
}

func (self *Model) RunModelContext(ctx context.Context, db *sql.DB, action string, ARGS interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error) {
    obj := self.GetAction(action)
    if obj == nil {
        return nil, fmt.Errorf("actions or action %s is nil", action)
    }
	if ARGS == nil {
		return obj.RunActionContext(ctx, db, &self.Table, nil, extra...)
	}

	switch t := ARGS.(type) {
	case map[string]interface{}:
		return obj.RunActionContext(ctx, db, &self.Table, t, extra...)
	case []map[string]interface{}:
		var data []map[string]interface{}
		for _, item := range t {
			lists, err := obj.RunActionContext(ctx, db, &self.Table, item, extra...)
			if err != nil {
				return nil, err
			}
			data = append(data, lists...)
		}
		return data, nil
	default:
		return nil, fmt.Errorf("wrong input data type: %#v", t)
	}

	return nil, nil
}
