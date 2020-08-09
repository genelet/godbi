package godbi

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"
)

func TestTable(t *testing.T) {
	str := `[
    {"name":"user_project", "alias":"j", "sortby":"c.componentid"},
    {"name":"user_component", "alias":"c", "type":"INNER", "using":"projectid"},
    {"name":"user_table", "alias":"t", "type":"LEFT", "using":"tableid"}]`
	tables := make([]*Join, 0)
	err := json.Unmarshal([]byte(str), &tables)
	if err != nil {
		panic(err)
	}

	if tables[0].Alias != `j` || tables[0].Sortby != `c.componentid` {
		t.Errorf("%v", tables[0])
	}
	if joinString(tables) != `user_project j
INNER JOIN user_component c USING (projectid)
LEFT JOIN user_table t USING (tableid)` {
		t.Errorf("===%s===", joinString(tables))
	}
}

func TestCrudStr(t *testing.T) {
	selectPar := "firstname"
	sql, labels, types := selectType(selectPar)
	if sql != "firstname" {
		t.Errorf("%s wanted", sql)
	}
	if labels[0] != "firstname" {
		t.Errorf("%s wanted", labels[0])
	}
	if types != nil {
		t.Errorf("%v wanted", types)
	}

	selectPars := []string{"firstname", "lastname", "id"}
	sql, labels, types = selectType(selectPars)
	if sql != "firstname, lastname, id" {
		t.Errorf("%s wanted", sql)
	}
	if labels[0] != "firstname" {
		t.Errorf("%s wanted", labels[0])
	}
	if types != nil {
		t.Errorf("%v wanted", types)
	}

	selectHash := map[string]string{"firstname": "First", "lastname": "Last", "id": "ID"}
	sql, labels, types = selectType(selectHash)
	if !strings.Contains(sql, "firstname") {
		t.Errorf("%s wanted", sql)
	}
	if types != nil {
		t.Errorf("%s wanted", types)
	}
	if !grep(labels, "First") {
		t.Errorf("%s wanted", labels)
	}

	extra := url.Values{}
	extra.Set("firstname", "Peter")
	sql, c := selectCondition(extra)
	if sql != "(firstname =?)" {
		t.Errorf("%s wanted", sql)
	}
	if c[0].(string) != "Peter" {
		t.Errorf("%s wanted", c[0].(string))
	}

	sql, c = selectCondition(extra, "user")
	if sql != "(user.firstname =?)" {
		t.Errorf("%s wanted", sql)
	}
	if c[0].(string) != "Peter" {
		t.Errorf("%s wanted", c[0].(string))
	}

	extra.Set("lastname", "Marcus")
	extra.Add("id", "1")
	extra.Add("id", "2")
	extra.Add("id", "3")
	extra.Add("id", "4")
	sql, c = selectCondition(extra)
	if !(strings.Contains(sql, "(firstname =?)") &&
		strings.Contains(sql, "(id IN (?,?,?,?))") &&
		strings.Contains(sql, "(lastname =?)")) {
		t.Errorf("%s wanted", sql)
	}
	if len(c) != 6 {
		t.Errorf("%v wanted", c)
	}

	crud := new(Table)
	crud.CurrentKeys = []string{"user_id", "edu_id"}
	ids := []interface{}{[]interface{}{11, 22}, []interface{}{33, 44, 55}}
	s, arr := crud.singleCondition(ids, extra)
	if !(strings.Contains(s, "user_id IN (?,?)") &&
		strings.Contains(s, "edu_id IN (?,?,?)") &&
		strings.Contains(s, "id IN (?,?,?,?)") &&
		strings.Contains(s, "(firstname =?)") &&
		strings.Contains(s, "(lastname =?)")) {
		t.Errorf("%s wanted", s)
	}
	if len(arr) != 11 {
		t.Errorf("%v wanted", ids)
		t.Errorf("%v wanted", extra)
		t.Errorf("%v wanted", s)
		t.Errorf("%v wanted", arr)
	}
}