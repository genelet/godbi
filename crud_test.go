package godbi

import (
	"testing"
)

func TestCrudDb(t *testing.T) {
	db, err := getdb()
	if err != nil {
		panic(err)
	}
	crud := newCrud(db, "atesting", "id", nil, nil)

	crud.execSQL(`drop table if exists atesting`)
	ret := crud.execSQL(`drop table if exists testing`)
	if ret != nil {
		t.Errorf("create table testing failed %s", ret.Error())
	}
	ret = crud.execSQL(`CREATE TABLE atesting (id int auto_increment, x varchar(255), y varchar(255), primary key (id))`)
	if ret != nil {
		t.Errorf("create table atesting failed")
	}
	hash := map[string]interface{}{"x":"a", "y":"b"}
	ret = crud.insertHash(hash)
	if crud.LastID != 1 {
		t.Errorf("%d wanted", crud.LastID)
	}
	hash["x"] = "c"
	hash["y"] = "d"
	ret = crud.insertHash(hash)
	id := crud.LastID
	if id != 2 {
		t.Errorf("%d wanted", id)
	}
	hash1 := map[string]interface{}{"y":"z"}
	ret = crud.updateHash(hash1, []interface{}{id})
	if ret != nil {
		t.Errorf("%s update table testing failed", ret.Error())
	}

	lists := make([]map[string]interface{}, 0)
	label := []string{"x", "y"}
	ret = crud.editHash(&lists, label, []interface{}{id})
	if ret != nil {
		t.Errorf("%s select table testing failed", ret.Error())
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
	ret = crud.topicsHash(&lists, label)
	if ret != nil {
		t.Errorf("%s select table testing failed", ret.Error())
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
	ret = crud.totalHash(&what)
	if ret != nil {
		t.Errorf("%s total table testing failed", ret.Error())
	}
	if what != 2 {
		t.Errorf("%d total table testing failed", what)
	}

	ret = crud.deleteHash(map[string]interface{}{"id": 1})
	if ret != nil {
		t.Errorf("%s delete table testing failed", ret.Error())
	}

	lists = make([]map[string]interface{}, 0)
	label = []string{"id", "x", "y"}
	ret = crud.topicsHash(&lists, label)
	if ret != nil {
		t.Errorf("%s select table testing failed", ret.Error())
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
	ret = crud.insertHash(hash)
	if ret.Error() == "" {
		t.Errorf("%s wanted", ret.Error())
	}

	hash1 = map[string]interface{}{"y": "zz"}
	ret = crud.updateHash(hash1, []interface{}{3})
	if ret != nil {
		t.Errorf("%s wanted", ret.Error())
	}
	if crud.Affected != 0 {
		t.Errorf("%d wanted", crud.Affected)
	}
	db.Close()
}
