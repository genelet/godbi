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
	RunActionContext(context.Context, *sql.DB, string, map[string]interface{}, ...map[string]interface{}) ([]map[string]interface{}, []*Page, error)
}

type Model struct {
	Table
	Actions map[string]interface{} `json:"actions,omitempty" hcl:"actions,optional"`
}

func NewModelJsonFile(fn string, custom ...map[string]Capability) (*Model, error) {
	dat, err := ioutil.ReadFile(fn)
	if err != nil { return nil, err }
	model := new(Model)
	err = json.Unmarshal(dat, &model)
	if err != nil { return nil, err }
	trans := make(map[string]interface{})
	for name, action := range model.Actions {
		jsonString, err := json.Marshal(action)
		if err != nil { return nil, err }
		var tran Capability
		switch name {
		case "insert": tran = new(Insert)
		case "update": tran = new(Update)
		case "insupd": tran = new(Insupd)
		case "edit":   tran = new(Edit)
		case "topics": tran = new(Topics)
		case "delete": tran = new(Delete)
		default:
			if custom != nil {
				if caction, ok := custom[0][name]; ok {
					tran = caction
				}
			}
		}
		err = json.Unmarshal(jsonString, &tran)
		if err != nil { return nil, err }
		tran.Fulfill(model.CurrentTable, model.Pks, model.IDAuto, model.Fks)
		trans[name] = tran
	}
	model.Actions = trans
	return model, nil
}

func (self *Model) RunContext(ctx context.Context, db *sql.DB, action string, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, []*Page, error) {
	if self.Actions == nil { return nil, nil, fmt.Errorf("no action assigned") }
	obi, ok := self.Actions[action]
	if !ok { return nil, nil, fmt.Errorf("action %s not found", action) }
	var tran Capability
	switch action {
	case "insert": tran = obi.(*Insert)
	case "update": tran = obi.(*Update)
	case "insupd": tran = obi.(*Insupd)
	case "edit":   tran = obi.(*Edit)
	case "topics": tran = obi.(*Topics)
	case "delete": tran = obi.(*Delete)
	default:
	}
	err := tran.CheckNull(ARGS)
	if err != nil { return nil, nil, err }
	lists, err := tran.RunContext(ctx, db, ARGS, extra...)
	if err != nil { return nil, nil, err }
	return lists, tran.GetNextpages(), nil
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
