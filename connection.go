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

func (self *Connection) FindExtra(item map[string]interface{}) map[string]interface{} {
	tableName := self.TableName
	if v, ok := self.RelateArgs[self.TableName]; ok {
		tableName = v
	}

	if v, ok := item[tableName]; ok {
		switch t := v.(type) {
		case map[string]interface{}:
			return t
		default:
		}
	}
	return nil
}

func (self *Connection) FindArgs(item interface{}) interface{} {
	if item == nil {
		return nil
	}

	tableName := self.TableName
	if v, ok := self.RelateArgs[self.TableName]; ok {
		tableName = v
	}

	switch t := item.(type) {
	case map[string]interface{}:
		if v, ok := t[tableName]; ok {
			switch s := v.(type) {
			case map[string]interface{}, []map[string]interface{}:
				return s
			default:
			}
		}
		return nil
	case []map[string]interface{}:
		var outs []map[string]interface{}
		for _, hash := range t {
			if v, ok := hash[tableName]; ok {
				switch s := v.(type) {
				case map[string]interface{}: // only map allowed here
					outs = append(outs, s)
				default:
				}
			}
		}
		return outs
	}
	return nil
}

// NextArg returns nextpage's args using current args map
//
func (self *Connection) NextArgs(item interface{}) interface{} {
	if item == nil {
		return nil
	}

	switch t := item.(type) {
	case map[string]interface{}:
		return createNextmap(self.RelateArgs, t)
	case []map[string]interface{}:
		var outs []map[string]interface{}
		for _, hash := range t {
			if x := createNextmap(self.RelateArgs, hash); x != nil {
				outs = append(outs, x)
			}
		}
		return outs
	}
	return nil
}

// NextExtra returns nextpage's extra using current extra map
//
func (self *Connection) NextExtra(args interface{}) map[string]interface{} {
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
				args = make(map[string]interface{})
			}
			switch t := u.(type) {
			case map[string]interface{}:
				for key, value := range t {
					args[key] = value
				}
			default:
				args[v] = t
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
func MergeArgs(args, items interface{}) interface{} {
	if args == nil {
		return items
	} else if items == nil {
		return args
	}

	switch t := items.(type) {
	case map[string]interface{}:
		return mergeMap(args, t)
	case []map[string]interface{}:
		var newArgs []map[string]interface{}
		for _, item := range t {
			newArgs = append(newArgs, mergeMap(args, item).(map[string]interface{}))
		}
		return newArgs
	default:
	}
	return args
}

func mergeMap(args interface{}, item map[string]interface{}) interface{} {
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
