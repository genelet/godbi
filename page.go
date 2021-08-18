package godbi

// Page type describes next page's structure
// Model: the name of the model
// Action: the method name on the model
// Manual: constraint conditions manually assigned
// RelateItem: current page's column versus next page's column, as constraints.
type Page struct {
	Model      string            `json:"model" hcl:"model,label"`
	Action     string            `json:"action" hcl:"action,label"`
	Manual     map[string]string `json:"manual,omitempty" hcl:"manual,optional"`
	RelateItem map[string]string `json:"relate_item,omitempty" hcl:"relate_item"`
}

func (self *Page) refresh(item, extra map[string]interface{}) (map[string]interface{}, bool) {
	newExtra := make(map[string]interface{})
	for k, v := range extra {
		newExtra[k] = v
	}
	found := false
	for k, v := range self.RelateItem {
		if t, ok := item[k]; ok {
			found = true
			newExtra[v] = t
			break
		}
	}
	return newExtra, found
}
