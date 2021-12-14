package godbi

import (
	"testing"
)

func TestGraphContext(t *testing.T) {
	ta, err := NewModelJsonFile("m_a.json")
	if err != nil {
		t.Fatal(err)
	}
	tb, err := NewModelJsonFile("m_b.json")
	if err != nil {
		t.Fatal(err)
	}
	graph := &Graph{Models:[]Navigate{ta, tb}}
	GraphGeneral(t, graph)
}

func TestGraphParse(t *testing.T) {
	graph, err := NewGraphJsonFile("graph.json")
	if err != nil {
		t.Fatal(err)
	}
	GraphGeneral(t, graph)
}

func TestGraphDelecs(t *testing.T) {
	graph, err := NewGraphJsonFile("graph2.json")
	if err != nil {
		t.Fatal(err)
	}
	db, ctx, METHODS := local2Vars()
	var lists []map[string]interface{}

	// the 1st web requests is assumed to create id=1 to the m_a and m_b tables:
	//
	args := map[string]interface{}{"x": "a1234567", "y": "b1234567", "z": "temp", "child": "john"}
    data2 := []map[string]interface{}{{"child": "john"}, {"child": "john2"}}
	graph.Initialize(map[string]interface{}{
		"m_a":map[string]interface{}{"insupd": args},
		"m_b":map[string]interface{}{"insert": data2},
	}, nil)
	if lists, err = graph.RunContext(ctx, db, "m_a", METHODS["PATCH"]); err != nil {
		panic(err)
	}
	if len(lists) != 1 {
		t.Errorf("%v", lists)
	}

	// the 2nd request just updates, becaues [x,y] is defined to the unique in ta.
	// but create a new record to tb for id=1, since insupd triggers insert in tb
	//
	args = map[string]interface{}{"x": "a1234567", "y": "b1234567", "z": "zzzzz"}
    data:= map[string]interface{}{"child": "sam"}
	graph.Initialize(map[string]interface{}{
		"m_a":map[string]interface{}{"insupd": args},
		"m_b":map[string]interface{}{"insert": data},
	}, nil)
	if lists, err = graph.RunContext(ctx, db, "m_a", METHODS["PATCH"]); err != nil {
		panic(err)
	}

	// the 3rd request creates id=2
	//
	args = map[string]interface{}{"x": "c1234567", "y": "d1234567", "z": "e1234"}
    data = map[string]interface{}{"child": "mary"}
	graph.Initialize(map[string]interface{}{
		"m_a":map[string]interface{}{"insert": args},
		"m_b":map[string]interface{}{"insert": data},
	}, nil)
	if lists, err = graph.RunContext(ctx, db, "m_a", METHODS["POST"]); err != nil {
		panic(err)
	}

	// the 4th request creates id=3
	//
	args = map[string]interface{}{"x": "e1234567", "y": "f1234567", "z": "e1234"}
    data = map[string]interface{}{"child": "marcus"}
	graph.Initialize(map[string]interface{}{
		"m_a":map[string]interface{}{"insert": args},
		"m_b":map[string]interface{}{"insert": data},
	}, nil)
	if lists, err = graph.RunContext(ctx, db, "m_a", METHODS["POST"]); err != nil {
		panic(err)
	}

	graph2Check(ctx, db, graph, METHODS, t)
}

func TestGraphDelecs2(t *testing.T) {
	graph, err := NewGraphJsonFile("graph2.json")
	if err != nil {
		t.Fatal(err)
	}
	db, ctx, METHODS := local2Vars()
	var lists []map[string]interface{}

	// the 1st web requests is assumed to create id=1 to the m_a and m_b tables:
	//
	args := map[string]interface{}{"x": "a1234567", "y": "b1234567", "z": "temp", "child": "john", "m_b": []map[string]interface{}{{"child": "john"}, {"child": "john2"}}}
	graph.Initialize(map[string]interface{}{
		"m_a":map[string]interface{}{"insupd": args},
	}, nil)
	if lists, err = graph.RunContext(ctx, db, "m_a", METHODS["PATCH"]); err != nil {
		panic(err)
	}
	if len(lists) != 1 {
		t.Errorf("%v", lists)
	}

	// the 2nd request just updates, becaues [x,y] is defined to the unique in ta.
	// but create a new record to tb for id=1, since insupd triggers insert in tb
	//
	args = map[string]interface{}{"x": "a1234567", "y": "b1234567", "z": "zzzzz", "m_b": map[string]interface{}{"child": "sam"}}
	graph.Initialize(map[string]interface{}{
		"m_a":map[string]interface{}{"insupd": args},
	}, nil)
	if lists, err = graph.RunContext(ctx, db, "m_a", METHODS["PATCH"]); err != nil {
		panic(err)
	}

	// the 3rd request creates id=2
	//
	args = map[string]interface{}{"x": "c1234567", "y": "d1234567", "z": "e1234", "m_b": map[string]interface{}{"child": "mary"}}
	graph.Initialize(map[string]interface{}{
		"m_a":map[string]interface{}{"insert": args},
	}, nil)
	if lists, err = graph.RunContext(ctx, db, "m_a", METHODS["POST"]); err != nil {
		panic(err)
	}

	// the 4th request creates id=3
	//
	args = map[string]interface{}{"x": "e1234567", "y": "f1234567", "z": "e1234", "m_b": map[string]interface{}{"child": "marcus"}}
	graph.Initialize(map[string]interface{}{
		"m_a":map[string]interface{}{"insert": args},
	}, nil)
	if lists, err = graph.RunContext(ctx, db, "m_a", METHODS["POST"]); err != nil {
		panic(err)
	}

	graph2Check(ctx, db, graph, METHODS, t)
}

func TestGraphThreeTables(t *testing.T) {
	graph, err := NewGraphJsonFile("graph3.json")
	if err != nil {
		t.Fatal(err)
	}
	GraphThreeGeneral(graph, t)
}

func TestGraphThreeTables2(t *testing.T) {
	graph, err := NewGraphJsonFile("graph31.json")
	if err != nil {
		t.Fatal(err)
	}
	GraphThreeGeneral(graph, t)
}
