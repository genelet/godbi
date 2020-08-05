package godbi

import (
    "testing"
    "net/url"
)

func TestPage(t *testing.T) {
	db, err := getdb()
	if err != nil {
		panic(err)
	}

    model, err := NewOne("m1.json")
    if err != nil { panic(err) }
    model.DB = db

	err = model.ExecSQL(`drop table if exists atesting`)
	if err != nil { panic(err) }
	err = model.ExecSQL(`CREATE TABLE atesting (id int auto_increment not null primary key, x varchar(8), y varchar(8), z varchar(8))`)
	if err != nil { panic(err) }

    hash := url.Values{"x":[]string{"a1234567"},"y":[]string{"b1234567"}}
    model.ARGS = hash
    err = model.Insupd()
    if err != nil { panic(err) }

    hash = url.Values{"x":[]string{"c1234567"},"y":[]string{"d1234567"},"z":[]string{"e1234"}}
    model.ARGS = hash
    err = model.Insupd()
    if err != nil { panic(err) }

    hash = url.Values{"x":[]string{"e1234567"},"y":[]string{"f1234567"},"z":[]string{"e1234"}}
    model.ARGS = hash
    err = model.Insupd()
    if err != nil { panic(err) }

	supp, err := NewTwo("m3.json")
    if err != nil { panic(err) }
    supp.DB = db
    err = supp.ExecSQL(`drop table if exists testing`)
    if err != nil { panic(err) }
    err = supp.ExecSQL(`CREATE TABLE testing (tid int auto_increment not null primary key, child varchar(8), id int)`)
    if err != nil { panic(err) }

    hash = url.Values{"id":[]string{"1"},"child":[]string{"john"}}
    supp.ARGS = hash
    err = supp.Insert()
    if err != nil { panic(err) }

    hash = url.Values{"id":[]string{"1"},"child":[]string{"sam"}}
    supp.ARGS = hash
    err = supp.Insert()
    if err != nil { panic(err) }

    hash = url.Values{"id":[]string{"2"},"child":[]string{"mary"}}
    supp.ARGS = hash
    err = supp.Insert()
    if err != nil { panic(err) }

    hash = url.Values{"id":[]string{"3"},"child":[]string{"kkk"}}
    supp.ARGS = hash
    err = supp.Insert()
    if err != nil { panic(err) }

	st, err := NewTwo("m3.json")
    if err != nil { panic(err) }

	models := make(map[string]Navigate)
	models["testing"] = st

	tt := make(map[string]interface{})
	tt["topics"] = func(args ...url.Values) error {
        return st.Topics(args...)
    }
	tt["customtwo"] = func(args ...url.Values) error {
        return st.Customtwo(args...)
    }
	actions := make(map[string]map[string]interface{})
	actions["testing"] = tt

	schema := &Schema{Models:models, Actions:actions}
	model.Scheme = schema

	item := make(map[string]interface{})
	page := &Page{Model:"testing", Action:"topics"}
	err = model.CallOnce(item, page)
	if err != nil { panic(err) }
	lists := item["testing_topics"].([]map[string]interface{})
// map[string]interface {}{"testing_topics":[]map[string]interface {}{map[string]interface {}{"child":"john", "id":1, "tid":1}, map[string]interface {}{"child":"sam", "id":1, "tid":2}, map[string]interface {}{"child":"mary", "id":2, "tid":3}, map[string]interface {}{"child":"kkk", "id":3, "tid":4}}}
	list0 := lists[0]
    if len(lists) != 4 || list0["child"].(string) != "john" {
		t.Errorf("%#v", list0)
	}

	item = make(map[string]interface{})
	page = &Page{Model:"testing", Action:"customtwo"}
	err = model.CallOnce(item, page)
	if err != nil { panic(err) }
	lists = item["testing_customtwo"].([]map[string]interface{})
	list0 = lists[0]
    if len(lists) != 1 || list0["four"].(int64) != 4 {
		t.Errorf("%#v", list0)
	}

	db.Close()
}

// using inherited structs, not the original model
type modelOne struct {
	Model
}
func NewOne(filename string) (*modelOne, error) {
	model, err := NewModel(filename)
    if err != nil { return nil, err }
	return &modelOne{*model}, nil
}
func (self *modelOne) Customone(extra ...url.Values) error {
	self.LISTS = make([]map[string]interface{},0)
	return self.SelectSQL(&self.LISTS,
"SELECT 1 AS one, 2 AS two")
}

// using inherited structs, not the original model
type modelTwo struct {
	Model
}
func NewTwo(filename string) (*modelTwo, error) {
	model, err := NewModel(filename)
    if err != nil { return nil, err }
	return &modelTwo{*model}, nil
}
func (self *modelTwo) Customtwo(extra ...url.Values) error {
	self.LISTS = make([]map[string]interface{},0)
	return self.SelectSQL(&self.LISTS,
"SELECT 3 AS three, 4 AS four")
}

func TestNavigate(t *testing.T) {
	db, err := getdb()
	if err != nil {
		panic(err)
	}

    model, err := NewOne("m1.json")
    if err != nil { panic(err) }
    model.DB = db

	err = model.ExecSQL(`drop table if exists atesting`)
	if err != nil { panic(err) }
	err = model.ExecSQL(`CREATE TABLE atesting (id int auto_increment not null primary key, x varchar(8), y varchar(8), z varchar(8))`)
	if err != nil { panic(err) }

    hash := url.Values{"x":[]string{"a1234567"},"y":[]string{"b1234567"}}
    model.ARGS = hash
    err = model.Insupd()
    if err != nil { panic(err) }

    hash = url.Values{"x":[]string{"c1234567"},"y":[]string{"d1234567"},"z":[]string{"e1234"}}
    model.ARGS = hash
    err = model.Insupd()
    if err != nil { panic(err) }

    hash = url.Values{"x":[]string{"e1234567"},"y":[]string{"f1234567"},"z":[]string{"e1234"}}
    model.ARGS = hash
    err = model.Insupd()
    if err != nil { panic(err) }

	err = model.Topics()
    if err != nil { panic(err) }
    lists := model.GetLists()
	list0 := lists[0]
    if len(lists) != 3 || list0["x"].(string) != "a1234567" {
		t.Errorf("%#v", list0)
	}

	err = model.Customone()
    if err != nil { panic(err) }
    lists = model.GetLists()
	list0 = lists[0]
    if len(lists) != 1 || list0["one"].(int64) != 1 {
		t.Errorf("%#v", list0)
	}

	db.Close()
}

func TestSchema(t *testing.T) {
	db, err := getdb()
	if err != nil {
		panic(err)
	}

    model, err := NewOne("m2.json")
    if err != nil { panic(err) }
    model.DB = db

	err = model.ExecSQL(`drop table if exists atesting`)
	if err != nil { panic(err) }
	err = model.ExecSQL(`CREATE TABLE atesting (id int auto_increment not null primary key, x varchar(8), y varchar(8), z varchar(8))`)
	if err != nil { panic(err) }

    hash := url.Values{"x":[]string{"a1234567"},"y":[]string{"b1234567"}}
    model.ARGS = hash
    err = model.Insupd()
    if err != nil { panic(err) }

    hash = url.Values{"x":[]string{"c1234567"},"y":[]string{"d1234567"},"z":[]string{"e1234"}}
    model.ARGS = hash
    err = model.Insupd()
    if err != nil { panic(err) }

    hash = url.Values{"x":[]string{"e1234567"},"y":[]string{"f1234567"},"z":[]string{"e1234"}}
    model.ARGS = hash
    err = model.Insupd()
    if err != nil { panic(err) }

	supp, err := NewTwo("m3.json")
    if err != nil { panic(err) }
    supp.DB = db
    err = supp.ExecSQL(`drop table if exists testing`)
    if err != nil { panic(err) }
    err = supp.ExecSQL(`CREATE TABLE testing (tid int auto_increment not null primary key, child varchar(8), id int)`)
    if err != nil { panic(err) }

    hash = url.Values{"id":[]string{"1"},"child":[]string{"john"}}
    supp.ARGS = hash
    err = supp.Insert()
    if err != nil { panic(err) }

    hash = url.Values{"id":[]string{"1"},"child":[]string{"sam"}}
    supp.ARGS = hash
    err = supp.Insert()
    if err != nil { panic(err) }

    hash = url.Values{"id":[]string{"2"},"child":[]string{"mary"}}
    supp.ARGS = hash
    err = supp.Insert()
    if err != nil { panic(err) }

    hash = url.Values{"id":[]string{"3"},"child":[]string{"kkk"}}
    supp.ARGS = hash
    err = supp.Insert()
    if err != nil { panic(err) }

	st, err := NewTwo("m3.json")
    if err != nil { panic(err) }

	models := make(map[string]Navigate)
	models["testing"] = st

	tt := make(map[string]interface{})
	tt["topics"] = func(args ...url.Values) error {
        return st.Topics(args...)
    }
	tt["customtwo"] = func(args ...url.Values) error {
        return st.Customtwo(args...)
    }
	actions := make(map[string]map[string]interface{})
	actions["testing"] = tt

	schema := &Schema{Models:models, Actions:actions}
	model.Scheme = schema

	err = model.Topics()
	if err != nil { panic(err) }
	lists := model.LISTS
// [map[id:1 testing_topics:[map[child:john id:1 tid:1] map[child:sam id:1 tid:2]] x:a1234567 y:b1234567] map[id:2 testing_topics:[map[child:mary id:2 tid:3]] x:c1234567 y:d1234567 z:e1234] map[id:3 testing_topics:[map[child:kkk id:3 tid:4]] x:e1234567 y:f1234567 z:e1234]]
    list0 := lists[0]
    relate := list0["testing_topics"].([]map[string]interface{})
    if len(lists) != 3 ||
        list0["x"].(string) != "a1234567" ||
        len(relate) != 2 ||
        relate[0]["child"].(string) != "john" {
        t.Errorf("%#v", list0)
        t.Errorf("%#v", relate)
    }

	db.Close()
}
