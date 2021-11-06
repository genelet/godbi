package godbi

import (
	"encoding/json"
	"testing"
)

func TestJoin(t *testing.T) {
	str := `[
    {"name":"user_project", "alias":"j", "sortby":"c.componentid"},
    {"name":"user_component", "alias":"c", "type":"INNER", "using":"projectid"},
    {"name":"user_table", "alias":"t", "type":"LEFT", "using":"tableid"}]`
	joins := make([]*Join, 0)
	err := json.Unmarshal([]byte(str), &joins)
	if err != nil {
		t.Fatal(err)
	}

	if joins[0].Alias != `j` || joins[0].Sortby != `c.componentid` {
		t.Errorf("%v", joins[0])
	}
	if joinString(joins) != `user_project j
INNER JOIN user_component c USING (projectid)
LEFT JOIN user_table t USING (tableid)` {
		t.Errorf("===%s===", joinString(joins))
	}
}
