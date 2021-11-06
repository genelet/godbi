package godbi

import (
	"context"
	"testing"
)

func TestGraphContext(t *testing.T) {
	db, err := getdb()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	db.Exec(`drop table if exists m_a`)
	db.Exec(`CREATE TABLE m_a (id int auto_increment not null primary key, x varchar(8), y varchar(8), z varchar(8))`)
	db.Exec(`drop table if exists m_b`)
	db.Exec(`CREATE TABLE m_b (tid int auto_increment not null primary key, child varchar(8), id int)`)

	ta, err := NewModelJsonFile("m_a.json")
	if err != nil {
		t.Fatal(err)
	}
	tb, err := NewModelJsonFile("m_b.json")
	if err != nil {
		t.Fatal(err)
	}

	METHODS := map[string]string{"LIST": "topics", "GET": "edit", "POST": "insert", "PUT": "update", "PATCH": "insupd", "DELETE": "delete"}
	graph := NewGraph(db, map[string]Navigate{"ta": ta, "tb": tb})

	var lists []map[string]interface{}
	// the 1st web requests is assumed to create id=1 to the m_a and m_b tables:
	//
	args := map[string]interface{}{"x": "a1234567", "y": "b1234567", "z": "temp", "child": "john"}
	if lists, err = graph.RunContext(ctx, "ta", METHODS["PATCH"], args); err != nil {
		panic(err)
	}

	// the 2nd request just updates, becaues [x,y] is defined to the unique in ta.
	// but create a new record to tb for id=1, since insupd triggers insert in tb
	//
	args = map[string]interface{}{"x": "a1234567", "y": "b1234567", "z": "zzzzz", "child": "sam"}
	if lists, err = graph.RunContext(ctx, "ta", METHODS["PATCH"], args); err != nil {
		panic(err)
	}

	// the 3rd request creates id=2
	//
	args = map[string]interface{}{"x": "c1234567", "y": "d1234567", "z": "e1234", "child": "mary"}
	if lists, err = graph.RunContext(ctx, "ta", METHODS["POST"], args); err != nil {
		panic(err)
	}

	// the 4th request creates id=3
	//
	args = map[string]interface{}{"x": "e1234567", "y": "f1234567", "z": "e1234", "child": "marcus"}
	if lists, err = graph.RunContext(ctx, "ta", METHODS["POST"], args); err != nil {
		panic(err)
	}

	// GET all
	args = map[string]interface{}{}
	lists, err = graph.RunContext(ctx, "ta", METHODS["LIST"], args)
	if err != nil {
		panic(err)
	}
	e1 := lists[0]["ta_edit"].([]map[string]interface{})
	e2 := e1[0]["tb_topics"].([]map[string]interface{})
	if e2[0]["child"].(string) != "john" {
		t.Errorf("%v", lists)
	}
	// [map[id:1 ta_edit:[map[id:1 tb_topics:[map[child:john id:1 tid:1] map[child:sam id:1 tid:2]] x:a1234567 y:b1234567 z:zzzzz]] x:a1234567 y:b1234567 z:zzzzz] map[id:2 ta_edit:[map[id:2 tb_topics:[map[child:mary id:2 tid:3]] x:c1234567 y:d1234567 z:e1234]] x:c1234567 y:d1234567 z:e1234] map[id:3 ta_edit:[map[id:3 tb_topics:[map[child:marcus id:3 tid:4]] x:e1234567 y:f1234567 z:e1234]] x:e1234567 y:f1234567 z:e1234]]

	// GET one
	args = map[string]interface{}{"id": 1}
	lists, err = graph.RunContext(ctx, "ta", METHODS["GET"], args)
	if err != nil {
		panic(err)
	}
	e2 = lists[0]["tb_topics"].([]map[string]interface{})
	if e2[0]["child"].(string) != "john" {
		t.Errorf("%v", lists)
	}
	// [map[id:1 tb_topics:[map[child:john id:1 tid:1] map[child:sam id:1 tid:2]] x:a1234567 y:b1234567 z:zzzzz]]

	// DELETE
	extra := map[string]interface{}{"id": 1}
	if lists, err = graph.RunContext(ctx, "tb", METHODS["DELETE"], map[string]interface{}{"tid": 1}, extra); err != nil {
		panic(err)
	}
	if lists, err = graph.RunContext(ctx, "tb", METHODS["DELETE"], map[string]interface{}{"tid": 2}, extra); err != nil {
		panic(err)
	}
	if lists, err = graph.RunContext(ctx, "ta", METHODS["DELETE"], map[string]interface{}{"id": 1}); err != nil {
		panic(err)
	}

	// GET all
	args = map[string]interface{}{}
	lists, err = graph.RunContext(ctx, "ta", METHODS["LIST"], args)
	if err != nil {
		panic(err)
	}
	e1 = lists[0]["ta_edit"].([]map[string]interface{})
	e2 = e1[0]["tb_topics"].([]map[string]interface{})
	if e2[0]["child"].(string) != "mary" {
		t.Errorf("%v", lists)
	}
	// [map[id:2 ta_edit:[map[id:2 tb_topics:[map[child:mary id:2 tid:3]] x:c1234567 y:d1234567 z:e1234]] x:c1234567 y:d1234567 z:e1234] map[id:3 ta_edit:[map[id:3 tb_topics:[map[child:marcus id:3 tid:4]] x:e1234567 y:f1234567 z:e1234]] x:e1234567 y:f1234567 z:e1234]]

	db.Exec(`drop table if exists m_a`)
	db.Exec(`drop table if exists m_b`)
}
