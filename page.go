package godbi

type Page struct {
	Model	string                 `json:"model"`
	Action	string                 `json:"action"`
	Alias	string                 `json:"alias,omitempty"`
	Ignore	bool                   `json:"ignore,omitempty"`
	Manual	map[string]string      `json:"manual,omitempty"`
	RelateItem map[string]string   `json:"relate_item,omitempty"`
}
