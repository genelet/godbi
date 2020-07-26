package godbi

import (
	"testing"
	"encoding/json"
)

func TestTable(t *testing.T) {
	str := `[
    {"name":"user_project", "alias":"j", "sortby":"c.componentid"},
    {"name":"user_component", "alias":"c", "type":"INNER", "using":"projectid"},
    {"name":"user_table", "alias":"t", "type":"LEFT", "using":"tableid"}]`
	tables := make([]*Table, 0)
    err := json.Unmarshal([]byte(str), &tables)
    if err != nil { panic(err) }

	if tables[0].Alias != `j` || tables[0].Sortby != `c.componentid` {
		t.Errorf("%v", tables[0])
	}
	if TableString(tables) != `user_project j
INNER JOIN user_component c USING (projectid)
LEFT JOIN user_table t USING (tableid)` {
		t.Errorf("===%s===", TableString(tables))
	}
}
