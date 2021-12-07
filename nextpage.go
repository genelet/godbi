package godbi

// Nextpage describes linked page
//
type Nextpage struct {
	// TableName: the name of the table
	TableName  string            `json:"table" hcl:"table,label"`
	// ActionName: the action on the model
	ActionName string            `json:"action" hcl:"action,label"`
	// RelateArgs: map current page's columns to nextpage's columns as input
	RelateArgs map[string]string `json:"relateArgs,omitempty" hcl:"relateArgs"`
	// RelateExtra: it maps current page's columns to nextpage's columns (for Nextpages), or earlier page's columns (for Prepares) as constrains.
	RelateExtra map[string]string `json:"relateExtra,omitempty" hcl:"relateExtra"`
}

// Subname is the marker string used to store the output
func (self *Nextpage) Subname() string {
	return self.TableName + "_" + self.ActionName
}

// NextArg returns nextpage's args by taking current item
//
func (self *Nextpage) NextArgs(item map[string]interface{}) map[string]interface{} {
	return createNextmap(self.RelateArgs, item)
}

// AppendArg appends current item to the existing args
//
func (self *Nextpage) AppendArgs(args interface{}, item map[string]interface{}) interface{} {
	return appendArgs(args, self.NextArgs(item))
}

// NextExtra returns nextpage's extra by taking current item
//
func (self *Nextpage) NextExtra(item map[string]interface{}) map[string]interface{} {
	return createNextmap(self.RelateExtra, item)
}

// AppendExtra appends current item to the existing extra
//
func (self *Nextpage) AppendExtra(extra, item map[string]interface{}) map[string]interface{} {
	return appendExtra(extra, self.NextExtra(item))
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

// cloneExtra clones extra to a new extra
//
func cloneExtra(extra map[string]interface{}) map[string]interface{} {
	if extra == nil {
		return nil
	}
	newExtra := map[string]interface{}{}
	for k, v := range extra {
		newExtra[k] = v
	}
	return newExtra
}

// cloneArgs clones args to a new args, keeping proper data type
//
func cloneArgs(args interface{}) interface{} {
	if args == nil {
		return nil
	}
	switch t := args.(type) {
	case map[string]interface{}:
		return cloneExtra(t)
	case []map[string]interface{}:
		var newArgs []map[string]interface{}
		for _, each := range t {
			newArgs = append(newArgs, cloneExtra(each))
		}
		return newArgs
	default:
	}
	return nil
}

func appendExtra(extra, item map[string]interface{}) map[string]interface{} {
	if extra == nil {
		return item
	} else if item == nil {
		return extra
	}
	newExtra := cloneExtra(extra)
	for k, v := range item {
		newExtra[k] = v
	}
	return newExtra
}

func appendArgs(args interface{}, item map[string]interface{}) interface{} {
	if args == nil {
		return item
	} else if item == nil {
		return args
	}
	switch t := args.(type) {
	case map[string]interface{}:
		return appendExtra(t, item)
	case []map[string]interface{}:
		var newArgs []map[string]interface{}
		for _, each := range t {
			newArgs = append(newArgs, appendExtra(each, item))
		}
		return newArgs
	}
	return nil
}
