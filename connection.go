package godbi

type ConnectType int
const (
	CONNECTDefault ConnectType = iota
	CONNECTOne
	CONNECTArray
	CONNECTMap
	CONNECTMany
)

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
	Dimension  ConnectType        `json:"dimension,omitempty" hcl:"dimension,label"`
	Marker     string             `json:"marker,omitempty" hcl:"marker,label"`
}

// Subname is the marker string used to store the output
func (self *Connection) Subname() string {
	if self.Marker != "" {
		return self.Marker
	}
	return self.TableName + "_" + self.ActionName
}

// FindExtra returns the value if the input i.e. item contains 
// the current table name as key.
//
func (self *Connection) FindExtra(item map[string]interface{}) map[string]interface{} {
	marker := self.Marker
	if marker == "" { return nil }

	if v, ok := item[marker]; ok {
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

	marker := self.Marker
	if marker == "" { return nil, false }

	switch t := args.(type) {
	case map[string]interface{}: // in practice, only this data type exists
		if v, ok := t[marker]; ok {
			switch s := v.(type) {
			case map[string]interface{}:
				if self.Dimension==CONNECTMap {
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
						outs = append(outs, map[string]interface{}{marker:x})
					}
				}
				return outs, true
			default:
			}
			return nil, true
		}
	default:
	}

	return nil, true
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
		var outs []map[string]interface{}
		for _, hash := range t {
			if x := createNextmap(self.RelateArgs, hash); x != nil {
				outs = append(outs, x)
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

func (self *Connection) Shorten(lists []map[string]interface{}) interface{} {
	if self.Dimension == CONNECTDefault || self.Marker == "" {
		return lists
	}

	switch self.Dimension {
	case CONNECTMap:
		output := make(map[string]interface{})
		for _, single := range lists {
			key, value := mapEntryToPair(single)
			if key != "" {
				output[key] = value
			}
		}
		return output
	case CONNECTMany:
		var output []interface{}
		for _, single := range lists {
			output = append(output, manyEntry(self.Marker, single))
		}
		return output
	case CONNECTArray:
		var output []interface{}
		for _, single := range lists {
			output = append(output, single[self.Marker])
		}
		return output
	case CONNECTOne:
		return lists[0]
	default:
	}
	return lists
}

func mapEntryToPair(single map[string]interface{}) (string, interface{}) {
	var key string
	var value interface{}
	var object map[string]interface{}

	for _, a := range single {
		var c map[string]interface{}
		switch b := a.(type) {
		case []interface{}:
			c = b[0].(map[string]interface{})
		case []map[string]interface{}:
			c = b[0]
		default:
			continue
		}

		for d, e := range c {
			if d == "key" {
				key = e.(string)
				continue
			}
			if d == "value" {
				value = e
				continue
			}
			switch f := e.(type) {
			case []interface{}:
				object = f[0].(map[string]interface{})
			case []map[string]interface{}:
				object = f[0]
			default:
			}
		}
	}

	if value != nil {
		return key, value
	}
	return key, object
}

func manyEntry(leader string, single map[string]interface{}) map[string]interface{} {
	output := make(map[string]interface{})
	var higher map[string]interface{}

	for key, value := range single {
		if key == leader {
			higher = make(map[string]interface{})
			switch t := value.(type) {
			case []interface{}:
				for k, v := range t[0].(map[string]interface{}) {
					higher[k] = v
				}
			case []map[string]interface{}:
				for k, v := range t[0] {
					higher[k] = v
				}
			case map[string]interface{}:
				for k, v := range t {
					higher[k] = v
				}
			default:
				higher[key] = value
			}
		} else {
			output[key] = value
		}
	}

	if higher != nil {
		for k, v := range higher {
			output[k] = v
		}
	}

	return output
}
