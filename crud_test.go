package godbi

import (
	"net/url"
	"strings"
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

	crud := new(Crud)
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

func TestCrudDb(t *testing.T) {
	db, err := getdb()
	if err != nil {
		panic(err)
	}
	crud := NewCrud(db, "atesting", "id", nil, nil)

	crud.ExecSQL(`drop table if exists atesting`)
	ret := crud.ExecSQL(`drop table if exists testing`)
	if ret != nil {
		t.Errorf("create table testing failed %s", ret.Error())
	}
	ret = crud.ExecSQL(`CREATE TABLE atesting (id int auto_increment, x varchar(255), y varchar(255), primary key (id))`)
	if ret != nil {
		t.Errorf("create table atesting failed")
	}
	hash := make(url.Values)
	hash.Set("x", "a")
	hash.Set("y", "b")
	ret = crud.insertHash(hash)
	if crud.LastID != 1 {
		t.Errorf("%d wanted", crud.LastID)
	}
	hash.Set("x", "c")
	hash.Set("y", "d")
	ret = crud.insertHash(hash)
	id := crud.LastID
	if id != 2 {
		t.Errorf("%d wanted", id)
	}
	hash1 := make(url.Values)
	hash1.Set("y", "z")
	ret = crud.updateHash(hash1, []interface{}{id})
	if ret != nil {
		t.Errorf("%s update table testing failed", ret.Error())
	}

	lists := make([]map[string]interface{}, 0)
	label := []string{"x", "y"}
	ret = crud.editHash(&lists, label, []interface{}{id})
	if ret != nil {
		t.Errorf("%s select table testing failed", ret.Error())
	}
	if len(lists) != 1 {
		t.Errorf("%d records returned from edit", len(lists))
	}
	if lists[0]["x"].(string) != "c" {
		t.Errorf("%s c wanted", lists[0]["x"].(string))
	}
	if lists[0]["y"].(string) != "z" {
		t.Errorf("%s z wanted", string(lists[0]["y"].(string)))
	}

	lists = make([]map[string]interface{}, 0)
	ret = crud.topicsHash(&lists, label)
	if ret != nil {
		t.Errorf("%s select table testing failed", ret.Error())
	}
	if len(lists) != 2 {
		t.Errorf("%d records returned from edit, should be 2", len(lists))
	}
	if string(lists[0]["x"].(string)) != "a" {
		t.Errorf("%s a wanted", string(lists[0]["x"].(string)))
	}
	if string(lists[0]["y"].(string)) != "b" {
		t.Errorf("%s b wanted", string(lists[0]["y"].(string)))
	}
	if string(lists[1]["x"].(string)) != "c" {
		t.Errorf("%s c wanted", string(lists[1]["x"].(string)))
	}
	if string(lists[1]["y"].(string)) != "z" {
		t.Errorf("%s z wanted", string(lists[1]["y"].(string)))
	}

	what := 0
	ret = crud.totalHash(&what)
	if ret != nil {
		t.Errorf("%s total table testing failed", ret.Error())
	}
	if what != 2 {
		t.Errorf("%d total table testing failed", what)
	}

	ret = crud.deleteHash(url.Values{"id":[]string{"1"}})
	if ret != nil {
		t.Errorf("%s delete table testing failed", ret.Error())
	}

	lists = make([]map[string]interface{}, 0)
	label = []string{"id", "x", "y"}
	ret = crud.topicsHash(&lists, label)
	if ret != nil {
		t.Errorf("%s select table testing failed", ret.Error())
	}
	if len(lists) != 1 {
		t.Errorf("%d records returned from edit", len(lists))
	}
	if lists[0]["id"].(int64) != 2 {
		t.Errorf("%d 2 wanted", lists[0]["x"].(int32))
		t.Errorf("%v wanted", lists[0]["x"])
	}
	if string(lists[0]["x"].(string)) != "c" {
		t.Errorf("%s c wanted", string(lists[0]["x"].(string)))
	}
	if string(lists[0]["y"].(string)) != "z" {
		t.Errorf("%s z wanted", string(lists[0]["y"].(string)))
	}

	hash = make(url.Values)
	hash.Set("id", "2")
	hash.Set("x", "a")
	hash.Set("y", "b")
	ret = crud.insertHash(hash)
	if ret.Error() == "" {
		t.Errorf("%s wanted", ret.Error())
	}

	hash1 = make(url.Values)
	hash1.Set("y", "zz")
	ret = crud.updateHash(hash1, []interface{}{3})
	if ret != nil {
		t.Errorf("%s wanted", ret.Error())
	}
	if crud.Affected != 0 {
		t.Errorf("%d wanted", crud.Affected)
	}
	db.Close()
}
