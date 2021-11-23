package godbi

// Edge type describes next page's structure
// Model: the name of the model
// Action: the method name on the model
// Manual: constraint conditions manually assigned
// RelateItem: current page's column versus next page's column, as constraints.
type Edge struct {
	TableName  string             `json:"model" hcl:"model,label"`
	ActionName string             `json:"action" hcl:"action,label"`
	Manual map[string]interface{} `json:"manual,omitempty" hcl:"manual,optional"`
	RelateItem map[string]string  `json:"relateItem,omitempty" hcl:"relateItem"`
}

func (self *Edge) manualRefresh(extra interface{}) interface{} {
	if self.Manual == nil {
		return extra
	}

	if extra == nil {
		// we got extra as [] in inserts, insupds. so this would be ignored
		newExtra := make(map[string]interface{})
		for k, v := range self.Manual {
			newExtra[k] = v
		}
		return newExtra
	}

	switch t := extra.(type) {
	case map[string]interface{}:
		newExtra := make(map[string]interface{})
		for k, v := range t {
			newExtra[k] = v
		}
		for k, v := range self.Manual {
			newExtra[k] = v
		}
		return newExtra
	case []map[string]interface{}:
		var newExtra []map[string]interface{}
		for _, item := range t {
			for k, v := range self.Manual {
				item[k] = v
			}
			newExtra = append(newExtra, item)
		}
		return newExtra
	default:
	}

	return nil
}

func short(t map[string]interface{}, item map[string]interface{}, related map[string]string) (map[string]interface{}, bool) {
	newExtra := make(map[string]interface{})
	if t != nil {
		for k, v := range t {
			newExtra[k] = v
		}
	}
	found := false
	for k, v := range related {
		if u, ok := item[k]; ok {
			found = true
			newExtra[v] = u
			break
		}
	}
	return newExtra, found
}

func (self *Edge) refresh(item map[string]interface{}, extra interface{}) (interface{}, bool) {
	if extra == nil {
		newExtra, found := short(nil, item, self.RelateItem)
		return newExtra, found
	}

	switch t := extra.(type) {
	case map[string]interface{}:
		newExtra, found := short(t, item, self.RelateItem)
		return newExtra, found
	case []map[string]interface{}:
		var newExtra []map[string]interface{}
		for _, each := range t {
			unit, found := short(each, item, self.RelateItem)
			if found == false { return nil, false }
			newExtra = append(newExtra, unit)
		}
		return newExtra, true
	default:
	}

	return nil, false
}
