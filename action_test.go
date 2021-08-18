package godbi

import (
	"testing"
	"context"
)

func TestCrudDbContext(t *testing.T) {
	db, err := getdb()
	if err != nil {
		panic(err)
	}
	crud := new(Model)
	crud.DB = db
	crud.CurrentTable = "atesting"
	crud.CurrentKey = "id"

	ctx := context.Background()

	db.Exec(`drop table if exists atesting`)
	db.Exec(`drop table if exists testing`)
	err = crud.DoSQLContext(ctx, `CREATE TABLE atesting (id int auto_increment, x varchar(255), y varchar(255), primary key (id))`)
	if err != nil {
		t.Errorf("create table atesting failed %#v", err)
	}
	hash := map[string]interface{}{"x":"a", "y":"b"}
	err = crud.insertHashContext(ctx, hash)
	if crud.LastID != 1 {
		t.Errorf("%d wanted", crud.LastID)
	}
	hash["x"] = "c"
	hash["y"] = "d"
	err = crud.insertHashContext(ctx, hash)
	id := crud.LastID
	if id != 2 {
		t.Errorf("%d wanted", id)
	}
	hash1 := map[string]interface{}{"y":"z"}
	err = crud.updateHashContext(ctx, hash1, []interface{}{id})
	if err != nil {
		t.Errorf("%s update table testing failed", err.Error())
	}

	lists := make([]map[string]interface{}, 0)
	label := map[string][2]string{"x":[2]string{"x",""}, "y":[2]string{"y",""}}
	err = crud.editHashContext(ctx, &lists, label, []interface{}{id})
	if err != nil {
		t.Errorf("%s select table testing failed", err.Error())
	}
	if len(lists) != 1 {
		t.Errorf("%d records returned from edit", len(lists))
	}
	if lists[0]["x"].(string) != "c" {
		t.Errorf("%s c wanted", lists[0]["x"].(string))
	}
	if lists[0]["y"].(string) != "z" {
		t.Errorf("%s z wanted", string(lists[0]["y"].(string)))
	}

	lists = make([]map[string]interface{}, 0)
	err = crud.topicsHashContext(ctx, &lists, label)
	if err != nil {
		t.Errorf("%s select table testing failed", err.Error())
	}
	if len(lists) != 2 {
		t.Errorf("%d records returned from edit, should be 2", len(lists))
	}
	if string(lists[0]["x"].(string)) != "a" {
		t.Errorf("%s a wanted", string(lists[0]["x"].(string)))
	}
	if string(lists[0]["y"].(string)) != "b" {
		t.Errorf("%s b wanted", string(lists[0]["y"].(string)))
	}
	if string(lists[1]["x"].(string)) != "c" {
		t.Errorf("%s c wanted", string(lists[1]["x"].(string)))
	}
	if string(lists[1]["y"].(string)) != "z" {
		t.Errorf("%s z wanted", string(lists[1]["y"].(string)))
	}

	what := 0
	err = crud.totalHashContext(ctx, &what)
	if err != nil {
		t.Errorf("%s total table testing failed", err.Error())
	}
	if what != 2 {
		t.Errorf("%d total table testing failed", what)
	}

	err = crud.deleteHashContext(ctx, map[string]interface{}{"id": 1})
	if err != nil {
		t.Errorf("%s delete table testing failed", err.Error())
	}

	lists = make([]map[string]interface{}, 0)
	label = map[string][2]string{"id":[2]string{"id",""}, "x":[2]string{"x",""}, "y":[2]string{"y",""}}
	err = crud.topicsHashContext(ctx, &lists, label)
	if err != nil {
		t.Errorf("%s select table testing failed", err.Error())
	}
	if len(lists) != 1 {
		t.Errorf("%d records returned from edit", len(lists))
	}
	if lists[0]["id"].(int64) != 2 {
		t.Errorf("%d 2 wanted", lists[0]["x"].(int32))
		t.Errorf("%v wanted", lists[0]["x"])
	}
	if string(lists[0]["x"].(string)) != "c" {
		t.Errorf("%s c wanted", string(lists[0]["x"].(string)))
	}
	if string(lists[0]["y"].(string)) != "z" {
		t.Errorf("%s z wanted", string(lists[0]["y"].(string)))
	}

	hash = map[string]interface{}{"id":"2", "x":"a", "y":"b"}
	err = crud.insertHashContext(ctx, hash)
	if err.Error() == "" {
		t.Errorf("%s wanted", err.Error())
	}

	hash1 = map[string]interface{}{"y": "zz"}
	err = crud.updateHashContext(ctx, hash1, []interface{}{3})
	if err != nil {
		t.Errorf("%s wanted", err.Error())
	}
	db.Close()
}
