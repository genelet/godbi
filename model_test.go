package godbi

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
)

type SQL struct {
	Action
	Columns   []string `json:"columns"`
	Statement string   `json:"statement"`
}

func (self *SQL) RunActionContext(ctx context.Context, db *sql.DB, t *Table, ARGS map[string]interface{}, extra ...interface{}) ([]map[string]interface{}, []*Nextpage, error) {
	v, ok := ARGS[self.Musts[0]]
	if !ok {
		return nil, nil, fmt.Errorf("missing %s in input", self.Musts[0])
	}
	lists := make([]map[string]interface{}, 0)
	dbi := &DBI{DB: db}
	err := dbi.SelectSQLContext(ctx, &lists, self.Statement, []interface{}{self.Columns}, v)
	return lists, self.Nextpages, err
}

func TestModel(t *testing.T) {
	custom := new(SQL)
	custom.ActionName = "sql"
	model, err := NewModelJsonFile("model.json", custom)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range model.Actions {
		k := v.GetName()
		switch k {
		case "topics":
			topics := v.(*Topics)
			if topics.Nextpages != nil {
				for i, page := range topics.Nextpages {
					if (i == 0 && (page.TableName != "adv_campaign")) ||
						(i == 1 && (page.RelateItem["campaign_id"] != "campaign_id")) {
						t.Errorf("%#v", page)
					}
				}
			}
		case "update":
			update := v.(*Update)
			if update.Columns[1] != "campaign_name" ||
				update.Empties[0] != "created" {
				t.Errorf("%#v", update)
			}
		case "sql":
			sql := v.(*SQL)
			if model.TableName != "adv_campaign" ||
				model.Pks[0] != "campaign_id" ||
				model.Fks[4] != "campaign_id_md5" ||
				sql.Nextpages[0].ActionName != "topics" ||
				sql.Statement != "SELECT x, y, z FROM a WHERE b=?" {
				t.Errorf("%#v", sql)
			}
		default:
		}
	}
}

func TestModelRun(t *testing.T) {
	db, err := getdb()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.Exec(`drop table if exists m_a`)
	db.Exec(`CREATE TABLE m_a (id int auto_increment not null primary key,
        x varchar(8), y varchar(8), z varchar(8))`)

	str := `{
    "table":"m_a",
    "pks":["id"],
    "id_auto":"id",
	"actions": [
	{
		"actionName": "insert",
		"musts":["x","y"],
		"columns":["x","y","z"]
	},
	{
		"actionName": "insupd",
		"uniques":["x","y"],
		"columns":["x","y","z"]
	},
	{
		"actionName": "delete",
		"musts":["id"]
	},
	{
		"actionName": "topics",
        "rename": [
            {"columnName":"x", "label":"x", "typeName":"string" },
            {"columnName":"y", "label":"y", "typeName":"string" },
            {"columnName":"z", "label":"z", "typeName":"string" },
            {"columnName":"id", "label":"id", "typeName":"int" }
        ]
	},
	{
		"actionName": "edit",
        "rename": [
            {"columnName":"x", "label":"x", "typeName":"string" },
            {"columnName":"y", "label":"y", "typeName":"string" },
            {"columnName":"z", "label":"z", "typeName":"string" },
            {"columnName":"id", "label":"id", "typeName":"int" }
		]
	}
]}`
	model, err := NewModelJson([]byte(str))
	if err != nil {
		t.Fatal(err)
	}

	var lists []map[string]interface{}
	var pages []*Nextpage
	// the 1st web requests is assumed to create id=1 to the m_a table
	//
	args := map[string]interface{}{"x": "a1234567", "y": "b1234567", "z": "temp", "child": "john"}
	lists, pages, err = model.RunModel(db, "insert", args)
	if err != nil {
		t.Fatal(err)
	}

	// the 2nd request just updates, becaues [x,y] is defined to the unique
	//
	args = map[string]interface{}{"x": "a1234567", "y": "b1234567", "z": "zzzzz", "child": "sam"}
	lists, pages, err = model.RunModel(db, "insupd", args)
	if err != nil {
		t.Fatal(err)
	}

	// the 3rd request creates id=2
	//
	args = map[string]interface{}{"x": "c1234567", "y": "d1234567", "z": "e1234", "child": "mary"}
	lists, pages, err = model.RunModel(db, "insert", args)
	if err != nil {
		t.Fatal(err)
	}

	// the 4th request creates id=3
	//
	args = map[string]interface{}{"x": "e1234567", "y": "f1234567", "z": "e1234", "child": "marcus"}
	lists, pages, err = model.RunModel(db, "insupd", args)
	if err != nil {
		t.Fatal(err)
	}

	// GET all
	args = map[string]interface{}{}
	lists, pages, err = model.RunModel(db, "topics", args)
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
	lists, pages, err = model.RunModel(db, "edit", args)
	if err != nil {
		t.Fatal(err)
	}
	e1 = lists[0]
	if len(lists) != 1 ||
		e1["id"].(int) != 1 ||
		e1["z"].(string) != "zzzzz" {
		t.Errorf("%v", lists)
		t.Errorf("%v", pages)
	}
	// [map[id:1 tb_topics:[map[child:john id:1 tid:1] map[child:sam id:1 tid:2]] x:a1234567 y:b1234567 z:zzzzz]]

	// DELETE
	args = map[string]interface{}{"id": 1}
	lists, pages, err = model.RunModel(db, "delete", args)
	if err != nil {
		t.Fatal(err)
	}

	// GET all
	args = map[string]interface{}{}
	lists, pages, err = model.RunModel(db, "topics", args)
	if err != nil {
		t.Fatal(err)
	}
	if len(lists) != 2 {
		t.Errorf("%v", lists)
		t.Errorf("%v", pages)
	}
	// [map[id:2 ta_edit:[map[id:2 tb_topics:[map[child:mary id:2 tid:3]] x:c1234567 y:d1234567 z:e1234]] x:c1234567 y:d1234567 z:e1234] map[id:3 ta_edit:[map[id:3 tb_topics:[map[child:marcus id:3 tid:4]] x:e1234567 y:f1234567 z:e1234]] x:e1234567 y:f1234567 z:e1234]]

	db.Exec(`drop table if exists m_a`)
	db.Exec(`drop table if exists m_b`)
}
