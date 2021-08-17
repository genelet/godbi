package godbi

import (
)

type Join struct {
    Name   string `json:"name" hcl:"name,label"`
    Alias  string `json:"alias,omitempty" hcl:"alias,optional"`
    Type   string `json:"type,omitempty" hcl:"type,optional"`
    Using  string `json:"using,omitempty" hcl:"using,optional"`
    On     string `json:"on,omitempty" hcl:"on,optional"`
    Sortby string `json:"sortby,omitempty" hcl:"sortby,optional"`
}

// joinString outputs the joined SQL statements from multiple tables.
//
func joinString(tables []*Join) string {
    sql := ""
    for i, table := range tables {
        name := table.Name
        if table.Alias != "" {
            name += " " + table.Alias
        }
        if i == 0 {
            sql = name
        } else if table.Using != "" {
            sql += "\n" + table.Type + " JOIN " + name + " USING (" + table.Using + ")"
        } else {
            sql += "\n" + table.Type + " JOIN " + name + " ON (" + table.On + ")"
        }
    }

    return sql
}

func (self *Join) getAlias() string {
    if self.Alias != "" {
        return self.Alias
    }
    return self.Name
}
