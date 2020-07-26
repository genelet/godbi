package godbi

// Page type describes next page's structure
// Model: the name of the model
// Action: the method name on the model
// Alias: the retrieved data is assigned to key: model_action as default. Use Alias to replace it.
// Ignore: if the key exists, don't run the next page
// Manual: constraint conditions manually assigned
// RelateItem: current page's column versus next page's column. The value is forced as constraint.
type Page struct {
	Model	string                 `json:"model"`
	Action	string                 `json:"action"`
	Alias	string                 `json:"alias,omitempty"`
	Ignore	bool                   `json:"ignore,omitempty"`
	Manual	map[string]string      `json:"manual,omitempty"`
	RelateItem map[string]string   `json:"relate_item,omitempty"`
}
