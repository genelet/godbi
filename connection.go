package godbi

// Connection describes linked page
// 1) for Nextpages, it maps item in lists to next ARGS and next Extra
// 2) for Prepares, it maps current ARGS to the next ARGS and next Extra
//
type Connection struct {
	// TableName: the name of the table
	TableName  string            `json:"tableName" hcl:"tableName,label"`

	// ActionName: the action on the model
	ActionName string            `json:"actionName" hcl:"actionName,label"`

	// RelateArgs: map current page's columns to nextpage's columns as input
	RelateArgs map[string]string `json:"relateArgs,omitempty" hcl:"relateArgs"`

	// RelateExtra: map current page's columns to nextpage's columns (for Nextpages), or prepared page's columns to current page's columns (for Prepares) as constrains.
	RelateExtra map[string]string `json:"relateExtra,omitempty" hcl:"relateExtra"`
}

// Subname is the marker string used to store the output
func (self *Connection) Subname() string {
	return self.TableName + "_" + self.ActionName
}

// NextArg returns nextpage's args by taking current item
//
func (self *Connection) NextArgs(item map[string]interface{}) map[string]interface{} {
	return createNextmap(self.RelateArgs, item)
}

// NextExtra returns nextpage's extra by taking current item
//
func (self *Connection) NextExtra(item map[string]interface{}) map[string]interface{} {
	return createNextmap(self.RelateExtra, item)
}

// PrepareArg returns prepare's args by taking current args
//
func (self *Connection) PrepareArgs(args interface{}) interface{} {
	if args == nil {
		return nil
	}

	switch t := args.(type) {
	case map[string]interface{}:
		return createNextmap(self.RelateArgs, t)
	case []map[string]interface{}:
		var outs []map[string]interface{}
		for _, item := range t {
			if x := createNextmap(self.RelateArgs, item); x != nil {
				outs = append(outs, x)
			}
		}
		return outs
	}
	return nil
}

// PrepareExtra returns prepare's extra by taking current args
//
func (self *Connection) PrepareExtra(args interface{}) map[string]interface{} {
	switch t := args.(type) {
	case map[string]interface{}:
	return createNextmap(self.RelateExtra, t)
	default:
	}
	return nil
}

func createNextmap(which map[string]string, item map[string]interface{}) map[string]interface{} {
	if which == nil {
		return nil
	}

	var args map[string]interface{}
	for k, v := range which {
		if u, ok := item[k]; ok {
			if args == nil {
				args = map[string]interface{}{v: u}
			} else {
				args[v] = u
			}
		}
	}
	return args
}

// CloneExtra clones extra to a new extra
//
func CloneExtra(extra map[string]interface{}) map[string]interface{} {
	if extra == nil {
		return nil
	}
	newExtra := map[string]interface{}{}
	for k, v := range extra {
		newExtra[k] = v
	}
	return newExtra
}

// CloneArgs clones args to a new args, keeping proper data type
//
func CloneArgs(args interface{}) interface{} {
	if args == nil {
		return nil
	}
	switch t := args.(type) {
	case map[string]interface{}:
		return CloneExtra(t)
	case []map[string]interface{}:
		var newArgs []map[string]interface{}
		for _, each := range t {
			newArgs = append(newArgs, CloneExtra(each))
		}
		return newArgs
	default:
	}
	return nil
}

// MergeExtra merges two maps
//
func MergeExtra(extra, item map[string]interface{}) map[string]interface{} {
	if extra == nil {
		return item
	} else if item == nil {
		return extra
	}
	newExtra := CloneExtra(extra)
	for k, v := range item {
		newExtra[k] = v
	}
	return newExtra
}

// MergeArgs merges map to either an existing map, or slice of map in which each element will be merged
//
func MergeArgs(args interface{}, item map[string]interface{}) interface{} {
	if args == nil {
		return item
	} else if item == nil {
		return args
	}
	switch t := args.(type) {
	case map[string]interface{}:
		return MergeExtra(t, item)
	case []map[string]interface{}:
		var newArgs []map[string]interface{}
		for _, each := range t {
			newArgs = append(newArgs, MergeExtra(each, item))
		}
		return newArgs
	}
	return nil
}
