package godbi

// Connection describes linked page
// 1) for Nextpages, it maps item in lists to next ARGS and next Extra
// 2) for Prepares, it maps current ARGS to the next ARGS and next Extra
//
type Connection struct {
	// TableName: the name of the table
	TableName  string             `json:"tableName" hcl:"tableName,label"`

	// ActionName: the action on the model
	ActionName string             `json:"actionName" hcl:"actionName,label"`

	// RelateArgs: map current page's columns to nextpage's columns as input
	RelateArgs map[string]string  `json:"relateArgs,omitempty" hcl:"relateArgs"`

	// RelateExtra: map current page's columns to nextpage's columns (for Nextpages), or prepared page's columns to current page's columns (for Prepares) as constrains.
	RelateExtra map[string]string `json:"relateExtra,omitempty" hcl:"relateExtra"`
	IsMapEntry bool               `json:"isMapEntry,omitempty" hcl:"isMapEntry,label"`
}

// Subname is the marker string used to store the output
func (self *Connection) Subname() string {
	return self.TableName + "_" + self.ActionName
}

// FindExtra returns the value if the input i.e. item contains 
// the current table name as key.
//
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

// FindArgs returns the value if the input i.e. args contains 
// the current table name as key.
//
func (self *Connection) FindArgs(args interface{}) (interface{}, bool) {
	if args == nil {
		return nil, true
	}

	topFound := false
	tableName := self.TableName
	if v, ok := self.RelateArgs[self.TableName]; ok {
		tableName = v
		topFound = true
	}

	switch t := args.(type) {
	case map[string]interface{}: // in practice, only this data type exists
		if v, ok := t[tableName]; ok {
			switch s := v.(type) {
			case map[string]interface{}:
				if self.IsMapEntry {
					var outs []map[string]interface{}
					for key, value := range s {
						outs = append(outs, map[string]interface{}{"key":key, "value":value})
					}
					return outs, true
				}
				return s, true
			case []map[string]interface{}:
				return s, true
			case []interface{}:
				var outs []map[string]interface{}
				for _, item := range s {
					switch x := item.(type) {
					case map[string]interface{}:
						outs = append(outs, x)
					default: // native types
						outs = append(outs, map[string]interface{}{tableName:x})
					}
				}
				return outs, true
			default:
			}
			return nil, true
		}
		return nil, topFound
	case []map[string]interface{}:
		var outs []map[string]interface{}
		found := false
		for _, hash := range t {
			if v, ok := hash[tableName]; ok {
				found = true
				switch s := v.(type) {
				case map[string]interface{}: // only map allowed here
					outs = append(outs, s)
				default:
				}
			}
		}
		return outs, found
	default:
	}
	return nil, false
}

// NextArg returns nextpage's args as the value of key  current args map
//
func (self *Connection) NextArgs(args interface{}) interface{} {
	if args == nil {
		return nil
	}
	if _, ok := self.RelateArgs["ALL"]; ok {
		return args
	}

	switch t := args.(type) {
	case map[string]interface{}:
		return createNextmap(self.RelateArgs, t)
	case []map[string]interface{}:
		var outs []interface{}
		for _, hash := range t {
			if x := createNextmap(self.RelateArgs, hash); x != nil {
				outs = append(outs, x)
			}
		}
		return outs
	case []interface{}:
		var outs []interface{}
		for _, hash := range t {
			if item, ok := hash.(map[string]interface{}); ok {
				if x := createNextmap(self.RelateArgs, item); x != nil {
					outs = append(outs, x)
				}
			}
		}
		return outs
	default:
	}
	return nil
}

// NextExtra returns nextpage's extra using current extra map
//
func (self *Connection) NextExtra(args interface{}) map[string]interface{} {
	if _, ok := self.RelateExtra["ALL"]; ok {
		if v, ok := args.(map[string]interface{}); ok {
			return v
		}
		return nil
	}

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
		var newArgs []interface{}
		for _, each := range t {
			newArgs = append(newArgs, CloneExtra(each))
		}
		return newArgs
	case []interface{}:
		var newArgs []interface{}
		for _, each := range t {
			if item, ok := each.(map[string]interface{}); ok {
				newArgs = append(newArgs, CloneExtra(item))
			}
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
		var newArgs []interface{}
		for _, item := range t {
			newArgs = append(newArgs, mergeMap(args, item).(map[string]interface{}))
		}
		return newArgs
	case []interface{}:
		var newArgs []interface{}
		for _, each := range t {
			if item, ok := each.(map[string]interface{}); ok {
				newArgs = append(newArgs, mergeMap(args, item).(map[string]interface{}))
			}
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
		var newArgs []interface{}
		for _, each := range t {
			newArgs = append(newArgs, MergeExtra(each, item))
		}
		return newArgs
	case []interface{}:
		var newArgs []interface{}
		for _, each := range t {
			if single, ok := each.(map[string]interface{}); ok {
				newArgs = append(newArgs, MergeExtra(single, item))
			}
		}
		return newArgs
	default:
	}
	return nil
}
