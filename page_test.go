package godbi

import (
    "testing"
    "net/url"
    "database/sql"
)

func TestPage(t *testing.T) {
	db, err := sql.Open("mysql", "eightran_goto:12pass34@/gotest")
	if err != nil {
		panic(err)
	}

    model, err := NewOne("m1.json")
    if err != nil { panic(err) }
    model.Db = db

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
    supp.Db = db
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

	methods := make(map[string]Restful)
	methods["testing"] = st

	tt := make(map[string]interface{})
	tt["topics"] = func(args ...url.Values) error {
        return st.Topics(args...)
    }
	tt["customtwo"] = func(args ...url.Values) error {
        return st.Customtwo(args...)
    }
	actions := make(map[string]map[string]interface{})
	actions["testing"] = tt

	schema := &Schema{Models:methods, Actions:actions}
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
