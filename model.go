package godbi

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
)

type Action interface {
	Fulfill(string, []string, string, []string)
	RunContext(context.Context, *sql.DB, map[string]interface{}, ...map[string]interface{}) ([]map[string]interface{}, error)
}

type Model struct {
	Component
	Actions map[string]interface{} `json:"actions,omitempty" hcl:"actions,optional"`
}

func NewModelJsonFile(fn string, custom ...map[string]Action) (*Model, error) {
	dat, err := ioutil.ReadFile(fn)
	if err != nil { return nil, err }
	model := new(Model)
	err = json.Unmarshal(dat, &model)
	if err != nil { return nil, err }
	trans := make(map[string]interface{})
	for name, action := range model.Actions {
		jsonString, err := json.Marshal(action)
		if err != nil { return nil, err }
		var tran Action
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
