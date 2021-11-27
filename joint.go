package godbi

type Joint struct {
	Name      string `json:"name" hcl:"name,label"`
	Alias     string `json:"alias,omitempty" hcl:"alias,optional"`
	JoinType  string `json:"type,omitempty" hcl:"type,optional"`
	JoinUsing string `json:"using,omitempty" hcl:"using,optional"`
	JoinOn    string `json:"on,omitempty" hcl:"on,optional"`
	Sortby    string `json:"sortby,omitempty" hcl:"sortby,optional"`
}

// joinString outputs the joined SQL statements from multiple tables.
//
func joinString(tables []*Joint) string {
	sql := ""
	for i, table := range tables {
		name := table.Name
		if table.Alias != "" {
			name += " " + table.Alias
		}
		if i == 0 {
			sql = name
		} else if table.JoinUsing != "" {
			sql += "\n" + table.JoinType + " JOIN " + name + " USING (" + table.JoinUsing + ")"
		} else {
			sql += "\n" + table.JoinType + " JOIN " + name + " ON (" + table.JoinOn + ")"
		}
	}

	return sql
}

func (self *Joint) getAlias() string {
	if self.Alias != "" {
		return self.Alias
	}
	return self.Name
}
