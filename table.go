package godbi

// Table describes a table used in multiple join select, which is represented as []*Table .
// Name: the name of the table
// Alias: the alias of the name
// Sortby: defines which column to sort. only used for the first table
// Using:  join by USING column name
// On: join by ON columns in the 2 tables
type Table struct {
	Name   string `json:"name"`
	Alias  string `json:"alias,omitempty"`
	Sortby string `json:"sortby,omitempty"`
	Type   string `json:"type,omitempty"`
	Using  string `json:"using,omitempty"`
	On     string `json:"on,omitempty"`
}

// TableString gives the joint SQL statements in place of the single table name.
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

func (self *Table) getAlias() string {
	if self.Alias != "" {
		return self.Alias
	}
	return self.Name
}
