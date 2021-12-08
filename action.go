package godbi

import (
	"context"
	"database/sql"
)

// Action is to implement Capability interface
//
type Capability interface {
	GetActionName() string
	GetPrepares()  []*Connection
	GetNextpages() []*Connection
	GetAppendix() interface{}
	RunActionContext(context.Context, *sql.DB, *Table, map[string]interface{}, ...map[string]interface{}) ([]map[string]interface{}, error)
}

type Action struct {
	ActionName string     `json:"actionName,omitempty" hcl:"actionName,optional"`
	Prepares  []*Connection `json:"Prepares,omitempty" hcl:"Prepares,block"`
	Nextpages []*Connection `json:"nextpages,omitempty" hcl:"nextpages,block"`
	Appendix  interface{} `json:"appendix,omitempty" hcl:"appendix,block"`
}

func (self *Action) GetActionName() string {
	return self.ActionName
}

func (self *Action) GetPrepares() []*Connection {
	return self.Prepares
}

func (self *Action) GetNextpages() []*Connection {
	return self.Nextpages
}

func (self *Action) GetAppendix() interface{} {
	return self.Appendix
}
