package godbi

type Table struct {
	Name   string    `json:"name"`
	Alias  string    `json:"alias,omitempty"`
	Sortby string    `json:"sortby,omitempty"`
	Type   string    `json:"type,omitempty"`
	Using  string    `json:"using,omitempty"`
	On     string    `json:"on,omitempty"`
}

func TableString(tables []*Table) string {
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

func (self *Table) GetAlias() string {
	if self.Alias != "" {
		return self.Alias
	}
	return self.Name
}
