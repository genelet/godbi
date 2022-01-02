package godbi

type Joint struct {
	TableName string `json:"tableName" hcl:"tableName,label"`
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
		name := table.TableName
		if table.Alias != "" {
			name += " " + table.Alias
		}
		joinType := "INNER"
		if table.JoinType != "" {
			joinType = table.JoinType
		}
		if i == 0 {
			sql = name
		} else if table.JoinUsing != "" {
			sql += "\n" + joinType + " JOIN " + name + " USING (" + table.JoinUsing + ")"
		} else {
			sql += "\n" + joinType + " JOIN " + name + " ON (" + table.JoinOn + ")"
		}
	}

	return sql
}

func (self *Joint) getAlias() string {
	if self.Alias != "" {
		return self.Alias
	}
	return self.TableName
}
