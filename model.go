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
	GetName() string
	NonePass(string) []string
	RunModelContext(context.Context, *sql.DB, string, map[string]interface{}, ...interface{}) ([]map[string]interface{}, []*Nextpage, error)
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

func NewModelJson(dat json.RawMessage, custom ...Capability) (*Model, error) {
	type m struct {
		Table
		Actions []interface{} `json:"actions,omitempty"`
	}
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
			if name==item.GetName() {
				tran = item
				found = true
				break
			}
		}
		if !found {
			switch name {
			case "insert":
				tran = new(Insert)
			case "inserts":
				tran = new(Inserts)
			case "update":
				tran = new(Update)
			case "insupd":
				tran = new(Insupd)
			case "insupds":
				tran = new(Insupds)
			case "edit":
				tran = new(Edit)
			case "topics":
				tran = new(Topics)
			case "delete":
				tran = new(Delete)
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

func (self *Model) GetName() string {
	return self.TableName
}

func (self *Model) GetAction(action string) Capability {
	if self.Actions != nil {
		for _, item := range self.Actions {
			if item.GetName() == action {
				return item
			}
		}
	}

	return nil
}

func (self *Model) RunModel(db *sql.DB, action string, ARGS map[string]interface{}, extra ...interface{}) ([]map[string]interface{}, []*Nextpage, error) {
	return self.RunModelContext(context.Background(), db, action, ARGS, extra...)
}

func (self *Model) RunModelContext(ctx context.Context, db *sql.DB, action string, ARGS map[string]interface{}, extra ...interface{}) ([]map[string]interface{}, []*Nextpage, error) {
	obj := self.GetAction(action)
	if obj == nil {
		return nil, nil, fmt.Errorf("actions or action %s is nil", action)
	}
	return obj.RunActionContext(ctx, db, &self.Table, ARGS, extra...)
}

func (self *Model) NonePass(action string) []string {
	if obj := self.GetAction(action); obj != nil {
		switch action {
		case "edit":
			return obj.(*Edit).defaultNames()
		case "topics":
			return obj.(*Topics).defaultNames()
		default:
		}
	}
	return nil
}
