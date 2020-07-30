package godbi

// Schema type: describes all models and actions in a schema
// Models: map between model name and model struct
// Actions: map between model name and actions, which is represented as a map between action name and action method
type Schema struct {
	Models  map[string]Restful
	Actions map[string]map[string]interface{}
}
