package godbi

import (
    "testing"
    "strings"
    "net/url"
)

func TestCrudStr(t *testing.T) {
	select_par :=  "firstname"
	sql, labels, types := selectType(select_par)
	if sql != "firstname" {
		t.Errorf("%s wanted", sql)
	}
	if labels[0] != "firstname" {
		t.Errorf("%s wanted", labels[0])
	}
	if types != nil {
		t.Errorf("%v wanted", types)
	}

	select_pars := []string{"firstname", "lastname", "id"}
	sql, labels, types = selectType(select_pars)
	if sql != "firstname, lastname, id" {
		t.Errorf("%s wanted", sql)
	}
	if labels[0] != "firstname" {
		t.Errorf("%s wanted", labels[0])
	}
	if types != nil {
		t.Errorf("%v wanted", types)
	}

	select_hash :=  map[string]string{"firstname":"First", "lastname":"Last", "id":"ID"}
	sql, labels, types = selectType(select_hash)
	if !strings.Contains(sql, "firstname") {
		t.Errorf("%s wanted", sql)
	}
	if !Grep(types, "First") {
		t.Errorf("%s wanted", types)
	}
	if !Grep(labels, "firstname") {
		t.Errorf("%s wanted", labels)
	}

	extra := url.Values{}
	extra.Set("firstname","Peter")
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

	extra.Set("lastname","Marcus")
	extra.Add("id","1")
	extra.Add("id","2")
	extra.Add("id","3")
	extra.Add("id","4")
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
	crud.CurrentKeys = []string{"user_id","edu_id"}
	ids := []interface{}{[]interface{}{11,22},[]interface{}{33,44,55}}
	s, arr := crud.singleCondition(ids, extra)
	if !( strings.Contains(s, "user_id IN (?,?)") &&
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
    if ret !=nil {
        t.Errorf("create table testing failed %s",ret.Error())
    }
	ret = crud.ExecSQL(`CREATE TABLE atesting (id int auto_increment, x varchar(255), y varchar(255), primary key (id))`)
	if ret !=nil {
		t.Errorf("create table atesting failed")
	}
	hash := make(url.Values)
	hash.Set("x","a")
	hash.Set("y","b")
	ret = crud.InsertHash(hash)
	if crud.LastId != 1 {
		t.Errorf("%d wanted", crud.LastId)
	}
	hash.Set("x","c")
	hash.Set("y","d")
	ret = crud.InsertHash(hash)
	id := crud.LastId
	if id != 2 {
		t.Errorf("%d wanted", id)
	}
	hash1 := make(url.Values)
	hash1.Set("y","z")
	ret = crud.UpdateHash(hash1, []interface{}{id})
	if ret !=nil {
		t.Errorf("%s update table testing failed", ret.Error())
	}

	lists := make([]map[string]interface{},0)
	label := []string{"x","y"}
	ret = crud.EditHash(&lists, label, []interface{}{id});
	if ret !=nil {
		t.Errorf("%s select table testing failed", ret.Error())
	}
	if len(lists)!=1 {
		t.Errorf("%d records returned from edit", len(lists))
	}
	if lists[0]["x"].(string) != "c" {
		t.Errorf("%s c wanted", lists[0]["x"].(string))
	}
	if lists[0]["y"].(string) != "z" {
		t.Errorf("%s z wanted", string(lists[0]["y"].(string)))
	}

	lists = make([]map[string]interface{},0)
	ret = crud.TopicsHash(&lists, label)
	if ret !=nil {
		t.Errorf("%s select table testing failed", ret.Error())
	}
	if len(lists)!=2 {
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
	ret = crud.TotalHash(&what)
	if ret !=nil {
		t.Errorf("%s total table testing failed", ret.Error())
	}
	if what !=2 {
		t.Errorf("%d total table testing failed", what)
	}

	ret = crud.DeleteHash([]interface{}{1})
	if ret !=nil {
		t.Errorf("%s delete table testing failed", ret.Error())
	}

	lists = make([]map[string]interface{},0)
	label = []string{"id","x","y"}
	ret = crud.TopicsHash(&lists, label)
	if ret !=nil {
		t.Errorf("%s select table testing failed", ret.Error())
	}
	if len(lists)!=1 {
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
	hash.Set("id","2")
	hash.Set("x","a")
	hash.Set("y","b")
	ret = crud.InsertHash(hash)
	if ret.Error() == "" {
		t.Errorf("%s wanted", ret.Error())
	}

	hash1 = make(url.Values)
	hash1.Set("y","zz")
	ret = crud.UpdateHash(hash1, []interface{}{3})
	if ret != nil {
		t.Errorf("%s wanted", ret.Error())
	}
	if crud.Affected != 0 {
		t.Errorf("%d wanted", crud.Affected)
	}
	db.Close()
}
