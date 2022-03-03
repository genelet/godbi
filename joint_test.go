package godbi

import (
	"encoding/json"
	"testing"
)

func TestJoint(t *testing.T) {
	str := `[
    {"tableName":"user_project", "alias":"j", "sortby":"c.componentid"},
    {"tableName":"user_component", "alias":"c", "type":"INNER", "using":"projectid"},
    {"tableName":"user_table", "alias":"t", "type":"LEFT", "using":"tableid"}]`
	joints := make([]*Joint, 0)
	err := json.Unmarshal([]byte(str), &joints)
	if err != nil {
		t.Fatal(err)
	}

	if joints[0].Alias != `j` || joints[0].Sortby != `c.componentid` {
		t.Errorf("%v", joints[0])
	}
	if joinString(joints) != `user_project j
INNER JOIN user_component c USING (projectid)
LEFT JOIN user_table t USING (tableid)` {
		t.Errorf("===%s===", joinString(joints))
	}
}
