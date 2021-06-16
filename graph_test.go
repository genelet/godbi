package godbi

import (
	"testing"
)

func TestGraph(t *testing.T) {
	db, err := getdb()
	if err != nil {
		panic(err)
	}

	model, err := NewModel("m2.json")
	if err != nil {
		panic(err)
	}
	model.DB = db

	err = model.execSQL(`drop table if exists atesting`)
	if err != nil {
		panic(err)
	}
	err = model.execSQL(`CREATE TABLE atesting (id int auto_increment not null primary key, x varchar(8), y varchar(8), z varchar(8))`)
	if err != nil {
		panic(err)
	}

	hash := map[string]interface{}{"x": "a1234567", "y": "b1234567"}
	model.SetDB(db)
	model.SetArgs(hash)
	err = model.Insupd()
	if err != nil {
		panic(err)
	}

	hash = map[string]interface{}{"x": "c1234567", "y": "d1234567", "z": "e1234"}
	model.SetArgs(hash)
	err = model.Insupd()
	if err != nil {
		panic(err)
	}

	hash = map[string]interface{}{"x": "e1234567", "y": "f1234567", "z": "e1234"}
	model.SetArgs(hash)
	err = model.Insupd()
	if err != nil {
		panic(err)
	}

	supp, err := NewModel("m3.json")
	if err != nil {
		panic(err)
	}
	supp.DB = db
	err = supp.execSQL(`drop table if exists testing`)
	if err != nil {
		panic(err)
	}
	err = supp.execSQL(`CREATE TABLE testing (tid int auto_increment not null primary key, child varchar(8), id int)`)
	if err != nil {
		panic(err)
	}

	hash = map[string]interface{}{"id": "1", "child": "john"}
	supp.SetArgs(hash)
	err = supp.Insert()
	if err != nil {
		panic(err)
	}

	hash = map[string]interface{}{"id": "1", "child": "sam"}
	supp.SetArgs(hash)
	err = supp.Insert()
	if err != nil {
		panic(err)
	}

	hash = map[string]interface{}{"id": "2", "child": "mary"}
	supp.SetArgs(hash)
	err = supp.Insert()
	if err != nil {
		panic(err)
	}

	hash = map[string]interface{}{"id": "3", "child": "kkk"}
	supp.SetArgs(hash)
	err = supp.Insert()
	if err != nil {
		panic(err)
	}

	st, err := NewModel("m3.json")
	if err != nil {
		panic(err)
	}

	graph := NewGraph(db, map[string]Navigate{"s": model, "testing": st})

	METHODS := map[string]string{"LIST":"topics", "GET":"edit", "POST":"insert", "PUT":"update", "PATCH":"insupd", "DELETE":"delete"}
	lists, err := graph.Run("s", METHODS["LIST"])
	if err != nil {
		panic(err)
	}
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

	lists, err = graph.Run("s", METHODS["LIST"], nil, nil, map[string]interface{}{"fields":[]string{"child","tid"}})
	if err != nil {
		panic(err)
	}
	// map[string]interface {}{"id":1, "testing_topics":[]map[string]interface {}{map[string]interface {}{"child":"john", "tid":1}, map[string]interface {}{"child":"sam", "tid":2}}, "x":"a1234567", "y":"b1234567"} 
	list0 = lists[0]
	relate = list0["testing_topics"].([]map[string]interface{})
	if _, ok := relate[0]["id"]; ok {
		t.Errorf("%#v", list0)
	}

	db.Close()
}
