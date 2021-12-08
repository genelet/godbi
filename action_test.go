package godbi

import (
	"encoding/json"
	"testing"
)

func TestAction(t *testing.T) {
	db, err := getdb()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.Exec(`drop table if exists m_a`)
	db.Exec(`CREATE TABLE m_a (id int auto_increment not null primary key,
        x varchar(8), y varchar(8), z varchar(8))`)

	tstr := `{
    "table":"m_a",
    "pks":["id"],
    "idAuto":"id",
	"rename":[
{"columnName":"x", "typeName":"string", "label":"x", "notnull": true},
{"columnName":"y", "typeName":"string", "label":"y", "notnull": true},
{"columnName":"z", "typeName":"string", "label":"z"},
{"columnName":"id","typeName":"int", "label":"id", "auto": true}
	],
	"uniques":["x","y"]
	}`
	table := new(Table)
	err = json.Unmarshal([]byte(tstr), table)
	if err != nil {
		t.Fatal(err)
	}

	insert := new(Insert)
	insupd := new(Insupd)
	topics := new(Topics)
	edit := new(Edit)
	dele := new(Delete)

	var lists []map[string]interface{}
	// the 1st web requests is assumed to create id=1 to the m_a table
	//
	args := map[string]interface{}{"x": "a1234567", "y": "b1234567", "z": "temp", "child": "john"}
	lists, err = insert.RunAction(db, table, args)
	if err != nil {
		t.Fatal(err)
	}

	// the 2nd request just updates, becaues [x,y] is defined to the unique
	//
	args = map[string]interface{}{"x": "a1234567", "y": "b1234567", "z": "zzzzz", "child": "sam"}
	lists, err = insupd.RunAction(db, table, args)
	if err != nil {
		t.Fatal(err)
	}

	// the 3rd request creates id=2
	//
	args = map[string]interface{}{"x": "c1234567", "y": "d1234567", "z": "e1234", "child": "mary"}
	lists, err = insert.RunAction(db, table, args)
	if err != nil {
		t.Fatal(err)
	}

	// the 4th request creates id=3
	//
	args = map[string]interface{}{"x": "e1234567", "y": "f1234567", "z": "e1234", "child": "marcus"}
	lists, err = insupd.RunAction(db, table, args)
	if err != nil {
		t.Fatal(err)
	}

	// GET all
	args = map[string]interface{}{}
	lists, err = topics.RunAction(db, table, args)
	if err != nil {
		t.Fatal(err)
	}
	// []map[string]interface {}{map[string]interface {}{"id":1, "x":"a1234567", "y":"b1234567", "z":"zzzzz"}, map[string]interface {}{"id":2, "x":"c1234567", "y":"d1234567", "z":"e1234"}, map[string]interface {}{"id":3, "x":"e1234567", "y":"f1234567", "z":"e1234"}}
	e1 := lists[0]
	e2 := lists[2]
	if len(lists) != 3 ||
		e1["id"].(int) != 1 ||
		e1["z"].(string) != "zzzzz" ||
		e2["y"].(string) != "f1234567" {
		t.Errorf("%v", lists)
	}

	// GET one
	args = map[string]interface{}{"id": 1}
	lists, err = edit.RunAction(db, table, args)
	if err != nil {
		t.Fatal(err)
	}
	e1 = lists[0]
	if len(lists) != 1 ||
		e1["id"].(int) != 1 ||
		e1["z"].(string) != "zzzzz" {
		t.Errorf("%v", lists)
	}

	// DELETE
	args = map[string]interface{}{"id": 1}
	lists, err = dele.RunAction(db, table, args)
	if err != nil {
		t.Fatal(err)
	}

	// GET all
	args = map[string]interface{}{}
	lists, err = topics.RunAction(db, table, args)
	if err != nil {
		t.Fatal(err)
	}
	if len(lists) != 2 {
		t.Errorf("%v", lists)
	}

	db.Exec(`drop table if exists m_a`)
	db.Exec(`drop table if exists m_b`)
}
