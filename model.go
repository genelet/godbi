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
	NonePass(string) []string
	RunModelContext(context.Context, *sql.DB, string, map[string]interface{}, ...map[string]interface{}) ([]map[string]interface{}, []*Edge, error)
}

type Model struct {
	Table
	Actions map[string]interface{} `json:"actions,omitempty" hcl:"actions,optional"`
}

func NewModelJsonFile(fn string, custom ...map[string]Capability) (*Model, error) {
	dat, err := ioutil.ReadFile(fn)
	if err != nil { return nil, err }
	return NewModelJson(dat, custom...)
}

func NewModelJson(dat []byte, custom ...map[string]Capability) (*Model, error) {
	model := new(Model)
	err := json.Unmarshal(dat, model)
	if err == nil {
		err = model.Assertion(custom...)
	}
	return model, err
}

func (self *Model) Assertion(custom ...map[string]Capability) error {
	trans := make(map[string]interface{})
	for name, action := range self.Actions {
		jsonString, err := json.Marshal(action)
		if err != nil { return err }
		var tran Capability
		if custom != nil && len(custom) > 0 && custom[0][name] != nil {
			tran = custom[0][name]
		} else {
			switch name {
			case "insert": tran = new(Insert)
			case "update": tran = new(Update)
			case "insupd": tran = new(Insupd)
			case "edit":   tran = new(Edit)
			case "topics": tran = new(Topics)
			case "delete": tran = new(Delete)
			default:
				return fmt.Errorf("action %s not defined", name)
			}
		}
		err = json.Unmarshal(jsonString, tran)
		if err != nil { return err }
		trans[name] = tran
	}
	self.Actions = trans
	return nil
}

func (self *Model) RunModel(db *sql.DB, action string, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, []*Edge, error) {
	return self.RunModelContext(context.Background(), db, action, ARGS, extra...)
}

func (self *Model) RunModelContext(ctx context.Context, db *sql.DB, action string, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, []*Edge, error) {
	if self.Actions == nil { return nil, nil, fmt.Errorf("actions is nil") }
	obi, ok := self.Actions[action]
	if !ok { return nil, nil, fmt.Errorf("action %s has no capability", action) }

	return obi.(Capability).RunActionContext(ctx, db, &self.Table, ARGS, extra...)
}

func (self *Model) NonePass(action string) []string {
	switch action {
	case "edit":
		if obi, ok := self.Actions["edit"]; ok {
			return obi.(*Edit).defaultNames()
		}
	case "topics":
		if obi, ok := self.Actions["topics"]; ok {
			return obi.(*Topics).defaultNames()
		}
	default:
	}
	return nil
}
