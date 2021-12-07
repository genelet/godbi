package godbi

// Nextpage type describes next page's structure
// Model: the name of the model
// Action: the method name on the model
// Manual: constraint conditions manually assigned
// RelateItem: current page's column versus next page's column, as constraints.
type Nextpage struct {
	TableName  string            `json:"table" hcl:"table,label"`
	ActionName string            `json:"action" hcl:"action,label"`
	RelateArgs map[string]string `json:"relateArgs,omitempty" hcl:"relateArgs"`
	RelateExtra map[string]string `json:"relateExtra,omitempty" hcl:"relateExtra"`
}

func (self *Nextpage) Subname() string {
	return self.TableName + "_" + self.ActionName
}

func (self *Nextpage) NextArgs(item map[string]interface{}) map[string]interface{} {
	return createNextmap(self.RelateArgs, item)
}

func (self *Nextpage) NextExtra(item map[string]interface{}) map[string]interface{} {
	return createNextmap(self.RelateExtra, item)
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

func cloneMap(extra map[string]interface{}) map[string]interface{} {
	if extra == nil {
		return nil
	}
	newExtra := map[string]interface{}{}
	for k, v := range extra {
		newExtra[k] = v
	}
	return newExtra
}

func cloneArgs(args interface{}) interface{} {
	if args == nil {
		return nil
	}
	switch t := args.(type) {
	case map[string]interface{}:
		return cloneMap(t)
	case []map[string]interface{}:
		var newArgs []map[string]interface{}
		for _, each := range t {
			newArgs = append(newArgs, cloneMap(each))
		}
		return newArgs
	default:
	}
	return nil
}

func appendMap(extra, item map[string]interface{}) map[string]interface{} {
	if extra == nil {
		return item
	} else if item == nil {
		return extra
	}
	newExtra := cloneMap(extra)
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
		return appendMap(t, item)
	case []map[string]interface{}:
		var newArgs []map[string]interface{}
		for _, each := range t {
			newArgs = append(newArgs, appendMap(each, item))
		}
		return newArgs
	}
	return nil
}

/*
// Use args as input, this function returns the output containing the
// manually assigned value
//
func (self *Nextpage) manualRefresh(args map[string]interface{}) map[string]interface{} {
	newArgs := map[string]interface{}{args}
	if self.Manual == nil {
		for k, v := range self.Manual {
			newArgs[k] = v
		}
	}
	return newArgs
}

// This function converts args to nextpage's args
//
// Only apply to "insert/insupd" where RelateArgs is defined for args
//
func (self *Nextpage) ownRefresh(args map[string]interface{}) map[string]interface{} {
	if args == nil {
		return nil
	}

	newArgs := map[string]interface{}{args}
	for k, v := range self.RelateArgs {
		if u, ok := args[k]; ok {
			delete(args, k)
			args[v] = u
		}
	}
	return newArgs
}

// This function adds more input on top of the existing input, 
// using item
//
func (self *Nextpage) argsRefresh(item map[string]interface{}, args interface{}) interface{} {
	newExtra := map[string]interface{}{}

	if args != nil {
		switch t := args.(type) {
		case []map[string]interface{}:
			var newArgs []map[string]interface{}
			for _, each := range t {
				hash := map[string]interface{}{each}
				for k, v := range self.RelateItem {
					if u, ok := item[k]; ok {
						hash[v] = u
					}
				}
				newArgs = append(newArgs, hash)
			}
			return newArgs
		case map[string]interface{}:
			newExtra = t
		default:
		}
	}

	for k, v := range self.RelateItem {
		if u, ok := item[k]; ok {
			newExtra[v] = u
		}
	}
	return newExtra
}

// This function adds more extra on top of the existing extra, 
// using item
//
func (self *Nextpage) itemRefresh(item, extra map[string]interface{}) interface{} {
	newExtra := map[string]interface{}{extra}
	for k, v := range self.RelateItem {
		if u, ok := item[k]; ok {
			newExtra[v] = u
		}
	}
	return newExtra
}

func appendMap(one, two map[string]interface{}) map[string]interface{} {
		if one == nil {
			return two
		} else if two == nil {
			return one
		}

		three := map[string]interface{}{one}
		for k, v := range two {
			three[k] = v
		}
		return three
}

func appendArgs(one interface{}, two map[string]interface{}) interface{} {
*/
