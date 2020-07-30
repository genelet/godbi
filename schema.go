package godbi

import (
	"net/url"
	"database/sql"
)

// Restful is interface whose methods have been implemented in Model
// GetLists: get the main data
// SetLists: setup the main data; pass nil will reset it to be nil
// UpdateModel: pass db handle, args and schema into Model
type Restful interface {
    GetLists() []map[string]interface{}
    SetLists([]map[string]interface{})
	UpdateModel(*sql.DB, url.Values, *Schema)
	CallOnce(map[string]interface{}, *Page, ...url.Values) error
}

// Schema describes all models and actions in a database schema
// Models: map between model name and model struct
// Actions: map between model name and actions,
// represented as a map between action name and action method
type Schema struct {
	Models  map[string]Restful
	Actions map[string]map[string]interface{}
}
