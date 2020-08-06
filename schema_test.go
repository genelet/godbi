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

    model, err := NewModel("m2.json")
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

	supp, err := NewModel("m3.json")
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


	st, err := NewModel("m3.json")
    if err != nil { panic(err) }

	methods := make(map[string]Navigate)
	methods["testing"] = st
	methods["s"] = model

	ss := make(map[string]interface{})
	ss["topics"] = func(args ...url.Values) error {
        return model.Topics(args...)
    }
	tt := make(map[string]interface{})
	tt["topics"] = func(args ...url.Values) error {
        return st.Topics(args...)
    }
	actions := make(map[string]map[string]interface{})
	actions["testing"] = tt
	actions["s"] = ss

	schema := &Schema{Models:methods, Actions:actions}

	lists, err := schema.Run("s","topics",url.Values{},db)
    if err != nil { panic(err) }
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
