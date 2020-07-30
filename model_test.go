package godbi

import (
	"math/rand"
    "testing"
    "strconv"
    "net/url"
    "database/sql"
)

func TestModelSimple(t *testing.T) {
	db, err := getdb()
	if err != nil {
		panic(err)
	}

    model := new(Model)
    model.Db = db
    model.CurrentTable = "testing"
    model.Sortby        ="sortby"
    model.Sortreverse   ="sortreverse"
    model.Pageno        ="pageno"
    model.Rowcount      ="rowcount"
    model.Totalno       ="totalno"
    model.Maxpageno     ="max_pageno"
    model.Fields        ="fields"
    model.Empties       ="empties"

	ret := model.ExecSQL(`drop table if exists testing`)
	if ret !=nil {
		t.Errorf("create table testing failed %s",ret.Error())
	}
	ret = model.ExecSQL(`CREATE TABLE testing (id int auto_increment, x varchar(255), y varchar(255), primary key (id))`)
	if ret !=nil {
		t.Errorf("create table testing failed %s",ret.Error())
	}

	args := make(url.Values)
	schema := new(Schema)
	model.UpdateModel(db, args, schema)

	model.CurrentKey = "id"
	model.CurrentIdAuto = "id"
	model.InsertPars = []string{"id","x","y"}
	model.TopicsPars = []string{"id","x","y"}

	args["x"] = []string{"a"}
	args["y"] = []string{"b"}
	ret = model.Insert()
	if model.LastId != 1 {
		t.Errorf("%d wanted", model.LastId)
	}
	hash := make(url.Values)
	hash.Set("x","c")
	hash.Set("y","d")
	ret = model.InsertHash(hash)
	id := model.LastId
	if id != 2 {
		t.Errorf("%d wanted", id)
	}

	err = model.Topics()
	if err != nil { panic(err) }
	LISTS := model.LISTS
	if len(LISTS) != 2 {
		t.Errorf("%d 2 columns wanted, %v", len(LISTS), LISTS)
	}

	model.UpdatePars = []string{"id","x","y"}
	model.EditPars   = []string{"id","x","y"}
	args.Set("id","2")
	args["x"] = []string{"c"}
	args["y"] = []string{"z"}
	ret = model.Update()
	if ret !=nil {
		t.Errorf("%s update table testing failed", ret.Error())
	}

	LISTS = model.LISTS
	ret = model.Edit()
	if ret !=nil {
		t.Errorf("%s edit table testing failed", ret.Error())
	}
	if len(LISTS)!=1 {
		t.Errorf("%d records returned from edit", len(LISTS))
	}
	if string(LISTS[0]["x"].(string)) != "c" {
		t.Errorf("%s c wanted", string(LISTS[0]["x"].(string)))
	}
	if string(LISTS[0]["y"].(string)) != "z" {
		t.Errorf("%s z wanted", string(LISTS[0]["y"].(string)))
	}

	ret = model.Topics()
	LISTS = model.LISTS
	if ret !=nil { panic(ret) }
	if len(LISTS)!=2 {
		t.Errorf("%d from topics, %v", len(LISTS), LISTS)
	}
	if string(LISTS[0]["x"].(string)) != "a" {
		t.Errorf("%s a wanted", string(LISTS[0]["x"].(string)))
	}
	if string(LISTS[0]["y"].(string)) != "b" {
		t.Errorf("%s b wanted", string(LISTS[0]["y"].(string)))
	}
	if string(LISTS[1]["x"].(string)) != "c" {
		t.Errorf("%s c wanted", string(LISTS[1]["x"].(string)))
	}
	if string(LISTS[1]["y"].(string)) != "z" {
		t.Errorf("%s z wanted", string(LISTS[1]["y"].(string)))
	}

	args["id"] = []string{"1"}
	ret = model.Delete()
	if ret !=nil {
		t.Errorf("%s delete table testing failed", ret.Error())
	}

	ret = model.Topics()
	LISTS = model.LISTS
	if ret !=nil {
		t.Errorf("%s select table testing failed", ret.Error())
	}
	if len(LISTS)!=1 {
		t.Errorf("%d records returned from edit", len(LISTS))
	}
	if LISTS[0]["id"].(int64) != 2 {
		t.Errorf("%d 2 wanted", LISTS[0]["x"].(int64))
		t.Errorf("%v wanted", LISTS[0]["x"])
	}
	if string(LISTS[0]["x"].(string)) != "c" {
		t.Errorf("%s c wanted", string(LISTS[0]["x"].(string)))
	}
	if string(LISTS[0]["y"].(string)) != "z" {
		t.Errorf("%s z wanted", string(LISTS[0]["y"].(string)))
	}

	args["id"] = []string{"2"}
	ret = model.Insert()
	if ret.Error() == "" {
		t.Errorf("%s wanted", ret.Error())
	}

	args["id"] = []string{"3"}
	args["y"] = []string{"zz"}
	ret = model.Update()
	if ret != nil || model.Affected != 0 {
		t.Errorf("%s %d wanted", ret.Error(), model.Affected)
	}

	model.ExecSQL(`truncate table testing`)
	delete(args,"id")
	for i:=1; i<100; i++ {
		delete(args, "id");
		args["x"] = []string{"a"}
		args["y"] = []string{"b"}
		ret = model.Insert()
		LISTS = model.LISTS
		if ret !=nil {
			t.Errorf("%s insert table testing failed", ret.Error())
		}
		if LISTS[0]["id"].(string) != strconv.Itoa(i) {
			t.Errorf("%d %s insert table auto id failed", i, LISTS[0]["id"].(string))
		}
	}

	for i :=1; i<100; i++ {
		args["id"] = []string{strconv.Itoa(i)}
		args["y"] = []string{"c"}
		ret = model.Update()
		LISTS = model.LISTS
		if ret !=nil {
			t.Errorf("%s update table testing failed", ret.Error())
		}
		if LISTS[0]["id"].(string) != strconv.Itoa(i) {
			t.Errorf("%d %s update id failed", i, LISTS[0]["id"].(string))
		}
		if LISTS[0]["y"].(string) != "c" {
			t.Errorf("%s update y failed", LISTS[0]["id"].(string))
		}
	}

	for i :=1; i<100; i++ {
		args["id"] = []string{strconv.Itoa(i)}
		ret = model.Edit()
		LISTS = model.LISTS
		if ret !=nil {
			t.Errorf("%s edit table testing failed", ret.Error())
		}
		if int(LISTS[0]["id"].(int64)) != i {
			t.Errorf("%d %d edit id failed", i, int(LISTS[0]["id"].(int64)))
		}
		if string(LISTS[0]["y"].(string)) != "c" {
			t.Errorf("%s edit y failed", string(LISTS[0]["id"].(string)))
		}
	}

	args["rowcount"] = []string{"20"}
	model.TotalForce = -1
	ret = model.Topics()
	LISTS = model.LISTS
	if ret !=nil {
		t.Errorf("%s edit table testing failed", ret.Error())
	}
	a := model.ARGS
	nt, err := strconv.Atoi(a["totalno"][0])
	if err != nil {panic(err)}
	nm, err := strconv.Atoi(a["max_pageno"][0])
	if err != nil {panic(err)}
	if nt!= 99 {
		t.Errorf("%d total is 99", nt)
	}
	if nm != 5 {
		t.Errorf("%d 5 pages", nm)
	}
	for i :=1; i<=20; i++ {
		if int(LISTS[i-1]["id"].(int64)) != i {
			t.Errorf("%d %d edit id failed", i, LISTS[i-1]["id"].(int64))
		}
	}

	args["pageno"] = []string{"3"}
	args["rowcount"] = []string{"20"}
	ret = model.Topics()
	LISTS = model.LISTS
	if ret !=nil {
		t.Errorf("%s topics table testing failed", ret.Error())
	}
	for i :=1; i<=20; i++ {
		if LISTS[i-1]["id"].(int64) != int64(40+i) {
			t.Errorf("%d %d topics id failed", 40+i, LISTS[i-1]["id"].(int))
		}
	}

	for i :=1; i<100; i++ {
		args["id"] = []string{strconv.Itoa(i)}
		ret = model.Delete()
		LISTS = model.LISTS
		if ret !=nil {
			t.Errorf("%s delete table testing failed", ret.Error())
		}
		x := LISTS[0]
		if x["id"].(string) != strconv.Itoa(i) {
			t.Errorf("%d %s delete id failed", i, x["id"].(string))
		}
	}
	db.Close()
}

func TestModel(t *testing.T) {
	db, err := sql.Open("mysql", "eightran_goto:12pass34@/gotest")
	if err != nil {
		panic(err)
	}

	model, err := NewModel("m1.json")
	if err != nil { panic(err) }
	ARGS := url.Values{}
	schema := new(Schema)
	model.UpdateModel(db, ARGS, schema)
	err = model.ExecSQL(`drop table if exists atesting`)
	if err != nil { panic(err) }
	err = model.ExecSQL(`CREATE TABLE atesting (id int auto_increment not null primary key, x varchar(8), y varchar(8), z varchar(8))`)
	if err != nil { panic(err) }

	str := model.OrderString()
	if str != "id" {
		t.Errorf("id expected, got %s", str)
	}

	ARGS.Set("sortreverse","1")
	ARGS.Set("rowcount","20")
	str = model.OrderString()
	if str != "id DESC LIMIT 20 OFFSET 0" {
		t.Errorf("'id DESC LIMIT 20 OFFSET 0' expected, got %s", str)
	}
	ARGS.Set("pageno","5")
	str = model.OrderString()
	if str != "id DESC LIMIT 20 OFFSET 80" {
		t.Errorf("'id DESC LIMIT 20 OFFSET 80' expected, got %s", str)
	}

	err = model.ExecSQL(`drop table if exists atesting`)
	if err != nil { panic(err) }
	err = model.ExecSQL(`CREATE TABLE atesting (id int auto_increment not null primary key, x varchar(8), y varchar(8), z varchar(8))`)
	if err != nil { panic(err) }

	for i:=0; i<100; i++ {
		hash := url.Values{}
		hash.Set("x","a1234567")
		hash.Set("y","b1234567")
		r := strconv.Itoa(int(rand.Int31()))
		if len(r)>8 { r=r[0:8] }
		hash.Set("z", r)
		model.ARGS = hash
		err = model.Insert()
		if err != nil { panic(err) }
	}
	model.ARGS.Set("rowcount","20")
	err = model.Topics()
	if err != nil { panic(err) }
    lists := model.LISTS
	if len(lists) !=20 {
		t.Errorf("%d records returned from topics", len(lists))
	}

	model.ARGS.Set("sortreverse","1")
	model.ARGS.Set("rowcount","20")
	model.ARGS.Set("pageno","5")
	str = model.OrderString()
	if str != "id DESC LIMIT 20 OFFSET 80" {
		t.Errorf("'id DESC LIMIT 20 OFFSET 80' expected, got %s", str)
	}
	if model.ARGS["totalno"][0] != "100" {
		t.Errorf("100 records expected, but %#v", model.ARGS)
	}
	db.Close()
}

func TestNextPages(t *testing.T) {
	db, err := sql.Open("mysql", "eightran_goto:12pass34@/gotest")
	if err != nil {
		panic(err)
	}

    model, err := NewModel("m2.json")
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

	supp, err := NewModel("m3.json")
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


	st, err := NewModel("m3.json")
    if err != nil { panic(err) }

	methods := make(map[string]Restful)
	methods["testing"] = st

	tt := make(map[string]interface{})
	tt["topics"] = func(args ...url.Values) error {
        return st.Topics(args...)
    }
	actions := make(map[string]map[string]interface{})
	actions["testing"] = tt

	schema := &Schema{Models:methods, Actions:actions}
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
