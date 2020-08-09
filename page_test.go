package godbi

import (
	"testing"
	"os"
    "net/url"
    "database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func TestPage(t *testing.T) {
	dbUser := os.Getenv("DBUSER")
	dbPass := os.Getenv("DBPASS")
	dbName := os.Getenv("DBNAME")
	db, err := sql.Open("mysql", dbUser + ":" + dbPass + "@/" + dbName)
	if err != nil { panic(err) }
	defer db.Close()

	db.Exec(`drop table if exists m_a`)
	db.Exec(`CREATE TABLE m_a (id int auto_increment not null primary key,
        x varchar(8), y varchar(8), z varchar(8))`)
    db.Exec(`drop table if exists m_b`)
    db.Exec(`CREATE TABLE m_b (tid int auto_increment not null primary key,
        child varchar(8), id int)`)

	ta, err := NewModel("m_a.json")
	if err != nil { panic(err) }
	tb, err := NewModel("m_b.json")
	if err != nil { panic(err) }

	// create action map for ta, the value of map is closure
    //
	action_ta := make(map[string]func(...url.Values)error)
	action_ta["topics"] = func(args ...url.Values) error { return ta.Topics(args...) }
	action_ta["insert"] = func(args ...url.Values) error { return ta.Insert(args...) }
	action_ta["insupd"] = func(args ...url.Values) error { return ta.Insupd(args...) }
	action_ta["delete"] = func(args ...url.Values) error { return ta.Delete(args...) }
	action_ta["edit"]   = func(args ...url.Values) error { return ta.Edit(args...) }
	ta.Actions = action_ta

	// create action map for ta, the value of map is closure
    //
	action_tb := make(map[string]func(...url.Values)error)
	action_tb["topics"] = func(args ...url.Values) error { return tb.Topics(args...) }
	action_tb["insert"] = func(args ...url.Values) error { return tb.Insert(args...) }
	action_tb["update"] = func(args ...url.Values) error { return tb.Update(args...) }
	action_tb["delete"] = func(args ...url.Values) error { return tb.Delete(args...) }
	action_tb["edit"]   = func(args ...url.Values) error { return tb.Edit(args...) }
	tb.Actions = action_tb

	schema := NewSchema(map[string]Navigate{"ta":ta, "tb":tb})
	schema.SetDB(db)

	methods := map[string]string{"GET":"topics", "GET_one":"edit", "POST":"insert", "PATCH":"insupd", "PUT":"update", "DELETE":"delete"}

	var lists []map[string]interface{}
    // the 1st web requests is assumed to create id=1 to the m_a and m_b tables:
    //
    args := url.Values{"x":[]string{"a1234567"},"y":[]string{"b1234567"},"z":[]string{"temp"}, "child":[]string{"john"}}
	if lists, err = schema.Run("ta", methods["PATCH"], args); err != nil { panic(err) }

    // the 2nd request just updates, becaues [x,y] is defined to the unique in ta.
    // but create a new record to tb for id=1, since insupd triggers insert in tb
    // 
    args = url.Values{"x":[]string{"a1234567"},"y":[]string{"b1234567"},"z":[]string{"zzzzz"}, "child":[]string{"sam"}}
	if lists, err = schema.Run("ta", methods["PATCH"], args); err != nil { panic(err) }

	// the 3rd request creates id=2
    //
    args = url.Values{"x":[]string{"c1234567"},"y":[]string{"d1234567"},"z":[]string{"e1234"},"child":[]string{"mary"}}
	if lists, err = schema.Run("ta", methods["POST"], args); err != nil { panic(err) }

	// the 4th request creates id=3
    //
    args = url.Values{"x":[]string{"e1234567"},"y":[]string{"f1234567"},"z":[]string{"e1234"},"child":[]string{"marcus"}}
	if lists, err = schema.Run("ta", methods["POST"], args); err != nil { panic(err) }

	// GET all
    args = url.Values{}
	lists, err = schema.Run("ta", methods["GET"], args)
	if err != nil { panic(err) }
	e1 := lists[0]["ta_edit"].([]map[string]interface{})
	e2 := e1[0]["tb_topics"].([]map[string]interface{})
	if e2[0]["child"].(string) != "john" {
		t.Errorf("%v", lists)
	}
// [map[id:1 ta_edit:[map[id:1 tb_topics:[map[child:john id:1 tid:1] map[child:sam id:1 tid:2]] x:a1234567 y:b1234567 z:zzzzz]] x:a1234567 y:b1234567 z:zzzzz] map[id:2 ta_edit:[map[id:2 tb_topics:[map[child:mary id:2 tid:3]] x:c1234567 y:d1234567 z:e1234]] x:c1234567 y:d1234567 z:e1234] map[id:3 ta_edit:[map[id:3 tb_topics:[map[child:marcus id:3 tid:4]] x:e1234567 y:f1234567 z:e1234]] x:e1234567 y:f1234567 z:e1234]]

	// GET one
    args = url.Values{"id":[]string{"1"}}
	lists, err = schema.Run("ta", methods["GET_one"], args)
	if err != nil { panic(err) }
	e2 = lists[0]["tb_topics"].([]map[string]interface{})
	if e2[0]["child"].(string) != "john" {
		t.Errorf("%v", lists)
	}
// [map[id:1 tb_topics:[map[child:john id:1 tid:1] map[child:sam id:1 tid:2]] x:a1234567 y:b1234567 z:zzzzz]]

	// DELETE
    extra := url.Values{"id":[]string{"1"}}
	if lists, err = schema.Run("tb", methods["DELETE"], url.Values{}, extra); err != nil { panic(err) }
	if lists, err = schema.Run("ta", methods["DELETE"], url.Values{}, extra); err != nil { panic(err) }

	// GET all
    args = url.Values{}
	lists, err = schema.Run("ta", methods["GET"], args)
	if err != nil { panic(err) }
	e1 = lists[0]["ta_edit"].([]map[string]interface{})
	e2 = e1[0]["tb_topics"].([]map[string]interface{})
	if e2[0]["child"].(string) != "mary" {
		t.Errorf("%v", lists)
	}
// [map[id:2 ta_edit:[map[id:2 tb_topics:[map[child:mary id:2 tid:3]] x:c1234567 y:d1234567 z:e1234]] x:c1234567 y:d1234567 z:e1234] map[id:3 ta_edit:[map[id:3 tb_topics:[map[child:marcus id:3 tid:4]] x:e1234567 y:f1234567 z:e1234]] x:e1234567 y:f1234567 z:e1234]]

	db.Exec(`drop table if exists m_a`)
	db.Exec(`drop table if exists m_b`)
}
