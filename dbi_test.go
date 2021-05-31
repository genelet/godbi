package godbi

import (
	"testing"
)

func TestProcedure(t *testing.T) {
	db, err := getdb()
	if err != nil {
		panic(err)
	}
	dbi := &DBI{DB:db}

	dbi.execSQL(`drop procedure if exists proc_w`)
	dbi.execSQL(`drop procedure if exists proc_w_resultset`)
	dbi.execSQL(`drop table if exists letters`)
	ret := dbi.execSQL(`create table letters(x varchar(1))`)
	if ret != nil {
		t.Errorf("create table failed")
	}

    ret = dbi.execSQL(`create procedure proc_w_resultset() begin insert into letters values('m'); insert into letters values('n'); select x from letters; select 1; select 2; insert into letters values('a'); end`)
    if ret != nil {
		t.Errorf("create stored procedure failed")
    }

	sql := `proc_w_resultset`
	lists := make([]map[string]interface{},0)
	ret = dbi.SelectProc(&lists, sql)
	if ret != nil {
		t.Errorf("Running select procedure failed")
	}
	if string(lists[0]["x"].(string)) != "m" {
		t.Errorf("%s m wanted", lists[0]["x"])
	}
	if string(lists[1]["x"].(string)) != "n" {
		t.Errorf("%s n wanted", lists[1]["x"])
	}

	ret = dbi.execSQL(`create procedure proc_w(IN x0 varchar(1),OUT y0 int) begin delete from letters; insert into letters values('m'); insert into letters values('n'); insert into letters values('p'); select x from letters where x=x0; insert into letters values('a'); set y0=100; end`)
	if ret != nil {
		t.Errorf("Running stored procedure failed")
	}

	sql = `proc_w`
	hash := make(map[string]interface{})
	lists = make([]map[string]interface{},0)
	ret = dbi.SelectDoProc(&lists, hash, []string{"y0"}, sql, "m")
	if ret != nil {
		t.Errorf("Running select do procedure failed")
	}
	if len(lists) != 1 {
		t.Errorf("%d returned", len(lists))
	}
	if string(lists[0]["x"].(string)) != "m" {
		t.Errorf("%s m wanted", lists[0]["x"])
	}
	if hash["y0"].(int64) != 100 {
		t.Errorf("%s 100 wanted", hash["y0"])
	}
	dbi.execSQL(`drop procedure if exists proc_w`)
	dbi.execSQL(`drop procedure if exists proc_w_resultset`)
	dbi.execSQL(`drop table if exists letters`)
	db.Close()
}
