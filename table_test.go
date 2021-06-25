package godbi

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTable1(t *testing.T) {
	str := `{
    "current_table": "m_a",
    "current_key" : "id",
    "current_id_auto" : "id",
    "insupd_pars" : ["x","y"],
    "insert_pars" : ["x","y","z"],
    "edit_pars"   : ["x","y","z","id"],
    "topics_pars" : [["id","int"],["x","string"],["y","string"]]}`
	table, err := newTable([]byte(str))
	if err != nil { t.Fatal(err) }

	x := table.editHashPars
	e := x["x"]
	if e[0] != "x" {
		t.Errorf("%#v", x)
	}
	y := table.topicsHashPars
	id:= y["id"]
	if id[0]!="id" || id[1]!="int" {
		t.Errorf("%#v", y)
	}
}

func TestTable2(t *testing.T) {
	str := `{
    "current_table": "m_a",
    "current_key" : "id",
    "current_id_auto" : "id",
    "insupd_pars" : ["x","y"],
    "insert_pars" : ["x","y","z"],
    "edit_pars"   : ["x","y","z","id"],
    "topics_pars" : [["id","int"],["x","string"],["y","string"]],
    "edit_hash" : {"x1":"x","y1":"y","z1":"z","id1":"id"},
    "topics_hash" : {"x1":["x","string"],"y1":["y","string"],"id1":["id","int"]} }`
	table, err := newTable([]byte(str))
	if err != nil { t.Fatal(err) }
	x := table.editHashPars
	x1 := x["x1"]
	if x1[0] != "x" {
		t.Errorf("%#v", x)
	}
	y := table.topicsHashPars
	id:= y["id1"]
	if id[0]!="id" || id[1]!="int" {
		t.Errorf("%#v", x)
	}
}

func TestTables(t *testing.T) {
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
	selectPar := generalHashPars(nil, []interface{}{"firstname"}, []string{"firstname"})
	sql, labels := selectType(selectPar)
	if sql != "firstname" {
		t.Errorf("%s wanted", sql)
	}
	f := labels[0].([2]string)
	if f[0] != "firstname" {
		t.Errorf("%s wanted", labels)
	}
	if f[1] != "" {
		t.Errorf("%v wanted", labels)
	}

	selectPars := generalHashPars(nil, []interface{}{"firstname", "lastname", "id"}, []string{"firstname", "lastname", "id"})
	sql, labels = selectType(selectPars)
	if !(strings.Contains(sql, "firstname") && strings.Contains(sql, "lastname") && strings.Contains(sql, "id")) {
		t.Errorf("%s wanted", sql)
	}
	ff := labels[0].([2]string)
	if !(ff[0] == "firstname" || ff[0] == "id" || ff[0] == "lastname") {
		t.Errorf("%s wanted", labels)
	}
	if ff[1] != "" {
		t.Errorf("%v wanted", labels)
	}

	selectHash := map[string][2]string{"firstname": [2]string{"First", "string"}, "lastname": [2]string{"Last", "string"}, "id": [2]string{"ID", "int"}}
	sql, labels = selectType(selectHash)
	if !strings.Contains(sql, "firstname") {
		t.Errorf("%s wanted", sql)
	}
	fff := labels[0].([2]string)
	str := fff[0]
	if !(str=="First" || str=="Last" || str=="ID")  {
		t.Errorf("%v wanted", labels)
	}

	extra := map[string]interface{}{"firstname": "Peter"}
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

	extra["lastname"] = "Marcus"
	extra["id"] = []string{"1","2","3","4"}
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
