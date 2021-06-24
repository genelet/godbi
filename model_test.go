package godbi

import (
	"math/rand"
	"context"
	"strconv"
	"testing"
)

func TestModelSimpleContext(t *testing.T) {
	db, err := getdb()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	model := new(Model)
	model.DB = db
	model.CurrentTable = "testing"
	model.Sortby = "sortby"
	model.Sortreverse = "sortreverse"
	model.Pageno = "pageno"
	model.Rowcount = "rowcount"
	model.Totalno = "totalno"
	model.Maxpageno = "max_pageno"
	model.Fields = "fields"

	db.Exec(`drop table if exists testing`)
	db.Exec(`CREATE TABLE testing (id int auto_increment, x varchar(255), y varchar(255), primary key (id))`)

	args := make(map[string]interface{})
	model.SetDB(db)
	model.SetArgs(args)

	model.CurrentKey = "id"
	model.CurrentIDAuto = "id"
	model.InsertPars = []string{"id", "x", "y"}
	model.TopicsPars = []interface{}{"id", "x", "y"}
	model.topicsHashPars = generalHashPars(nil, model.TopicsPars, nil)

	args["x"] = "a"
	args["y"] = "b"
	err = model.InsertContext(ctx)
	if err != nil { t.Fatal(err) }
	if model.LastID != 1 {
		t.Errorf("1 wanted but got %d", model.LastID)
	}
	hash := make(map[string]interface{})
	hash["x"] = "c"
	hash["y"] = "d"
	err = model.insertHashContext(ctx, hash)
	if err != nil { t.Fatal(err) }
	id := model.LastID
	if id != 2 {
		t.Errorf("2 wanted but got %d", id)
	}

	err = model.TopicsContext(ctx)
	if err != nil {
		panic(err)
	}
	LISTS := model.GetLists()
	if len(LISTS) != 2 {
		t.Errorf("%d 2 columns wanted, %v", len(LISTS), LISTS)
	}

	model.UpdatePars = []string{"id", "x", "y"}
	model.EditPars = []interface{}{"id", "x", "y"}
	model.editHashPars = generalHashPars(nil, model.EditPars, nil)
	args["id"] = 2
	args["x"] = "c"
	args["y"] = "z"
	err = model.UpdateContext(ctx)
	if err != nil { t.Fatal(err) }

	LISTS = model.GetLists()
	err = model.EditContext(ctx)
	if err != nil { t.Fatal(err) }
	if len(LISTS) != 1 {
		t.Errorf("%d records returned from edit", len(LISTS))
	}
	if string(LISTS[0]["x"].(string)) != "c" {
		t.Errorf("%s c wanted", string(LISTS[0]["x"].(string)))
	}
	if string(LISTS[0]["y"].(string)) != "z" {
		t.Errorf("%s z wanted", string(LISTS[0]["y"].(string)))
	}

	err = model.TopicsContext(ctx)
	if err != nil { t.Fatal(err) }
	LISTS = model.GetLists()
	if len(LISTS) != 2 {
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

	args["id"] = "1"
	err = model.DeleteContext(ctx, map[string]interface{}{"id": args["id"]})
	if err != nil { t.Fatal(err) }

	err = model.TopicsContext(ctx)
	if err != nil { t.Fatal(err) }
	LISTS = model.GetLists()
	if len(LISTS) != 1 {
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

	args["id"] = "2"
	err = model.InsertContext(ctx)
	if err == nil {
		t.Errorf("key id=2 exists. but still inserted. wrong")
	}

	args["id"] = "3"
	args["y"] = "zz"
	err = model.UpdateContext(ctx)
	if err != nil { t.Fatal(err) }

	db.Exec(`truncate table testing`)
	delete(args, "id")
	for i := 1; i < 100; i++ {
		delete(args, "id")
		args["x"] = "a"
		args["y"] = "b"
		err = model.InsertContext(ctx)
		if err != nil { t.Fatal(err) }
		LISTS = model.GetLists()
		if LISTS[0]["id"].(string) != strconv.Itoa(i) {
			t.Errorf("%d %s insert table auto id failed", i, LISTS[0]["id"].(string))
		}
	}

	for i := 1; i < 100; i++ {
		args["id"] = strconv.Itoa(i)
		args["y"] = "c"
		err = model.UpdateContext(ctx)
		if err != nil { t.Fatal(err) }
		LISTS = model.GetLists()
		if LISTS[0]["id"].(string) != strconv.Itoa(i) {
			t.Errorf("%d %s update id failed", i, LISTS[0]["id"].(string))
		}
		if LISTS[0]["y"].(string) != "c" {
			t.Errorf("%s update y failed", LISTS[0]["id"].(string))
		}
	}

	for i := 1; i < 100; i++ {
		args["id"] = strconv.Itoa(i)
		err = model.EditContext(ctx)
		if err != nil { t.Fatal(err) }
		LISTS = model.GetLists()
		if int(LISTS[0]["id"].(int64)) != i {
			t.Errorf("%d %d edit id failed", i, int(LISTS[0]["id"].(int64)))
		}
		if string(LISTS[0]["y"].(string)) != "c" {
			t.Errorf("%v edit y failed", LISTS[0])
		}
	}

	args["rowcount"] = 20
	model.TotalForce = -1
	err = model.TopicsContext(ctx)
	if err != nil { t.Fatal(err) }
	LISTS = model.GetLists()
	a := args
	nt := a["totalno"].(int)
	nm := a["max_pageno"].(int)
	if nt != 99 {
		t.Errorf("%d total is 99", nt)
	}
	if nm != 5 {
		t.Errorf("%d 5 pages", nm)
	}
	for i := 1; i <= 20; i++ {
		if int(LISTS[i-1]["id"].(int64)) != i {
			t.Errorf("%d %d edit id failed", i, LISTS[i-1]["id"].(int64))
		}
	}

	args["pageno"] = 3
	args["rowcount"] = 20
	err = model.TopicsContext(ctx)
	if err != nil { t.Fatal(err) }
	LISTS = model.GetLists()
	for i := 1; i <= 20; i++ {
		if LISTS[i-1]["id"].(int64) != int64(40+i) {
			t.Errorf("%d %d topics id failed", 40+i, LISTS[i-1]["id"].(int))
		}
	}

	for i := 1; i < 100; i++ {
		args["id"] = strconv.Itoa(i)
		err = model.DeleteContext(ctx, map[string]interface{}{"id": args["id"]})
		if err != nil { t.Fatal(err) }
		LISTS = model.GetLists()
		x := LISTS[0]
		if x["id"].(string) != strconv.FormatInt(int64(i), 10) {
			t.Errorf("%d %v delete id failed", i, x)
		}
	}
	db.Close()
}

func TestModelContext(t *testing.T) {
	db, err := getdb()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	model, err := NewModel("m1.json")
	if err != nil {
		panic(err)
	}
	ARGS := map[string]interface{}{}
	model.SetDB(db)
	model.SetArgs(ARGS)
	db.Exec(`drop table if exists atesting`)
	db.Exec(`CREATE TABLE atesting (id int auto_increment not null primary key, x varchar(8), y varchar(8), z varchar(8))`)
	if err != nil {
		panic(err)
	}

	str := model.orderString()
	if str != "ORDER BY id" {
		t.Errorf("id expected, got %s", str)
	}

	ARGS["sortreverse"] = "1"
	ARGS["rowcount"] = 20
	str = model.orderString()
	if str != "ORDER BY id DESC LIMIT 20 OFFSET 0" {
		t.Errorf("'id DESC LIMIT 20 OFFSET 0' expected, got %s", str)
	}
	ARGS["pageno"] = 5
	str = model.orderString()
	if str != "ORDER BY id DESC LIMIT 20 OFFSET 80" {
		t.Errorf("'ORDER BY id DESC LIMIT 20 OFFSET 80' expected, got %s", str)
	}

	db.Exec(`drop table if exists atesting`)
	db.Exec(`CREATE TABLE atesting (id int auto_increment not null primary key, x varchar(8), y varchar(8), z varchar(8))`)

	var hash map[string]interface{}
	for i := 0; i < 100; i++ {
		hash = map[string]interface{}{"x": "a1234567", "y": "b1234567"}
		r := strconv.Itoa(int(rand.Int31()))
		if len(r) > 8 {
			r = r[0:8]
		}
		hash["z"] = r
		model.SetArgs(hash)
		err = model.InsertContext(ctx)
		if err != nil {
			panic(err)
		}
	}
	a := hash
	a["rowcount"] = 20
	model.SetArgs(a)
	err = model.TopicsContext(ctx)
	if err != nil {
		panic(err)
	}
	lists := model.GetLists()
	if len(lists) != 20 {
		t.Errorf("%d records returned from topics", len(lists))
	}

	a["sortreverse"] = "1"
	a["rowcount"] = 20
	a["pageno"] = 5
	model.SetArgs(a)
	str = model.orderString()
	if str != "ORDER BY id DESC LIMIT 20 OFFSET 80" {
		t.Errorf("'ORDER BY id DESC LIMIT 20 OFFSET 80' expected, got %s", str)
	}
	if a["totalno"].(int) != 100 {
		t.Errorf("100 records expected, but %#v", a)
	}
	db.Close()
}
