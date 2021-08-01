package godbi

import (
	"context"
	"testing"
)

func TestContextProcedure(t *testing.T) {
	db, err := getdb()
	if err != nil {
		panic(err)
	}
	dbi := &DBI{DB:db}
	ctx := context.Background()

	dbi.Exec(`drop procedure if exists proc_w`)
	dbi.Exec(`drop procedure if exists proc_w_resultset`)
	dbi.Exec(`drop table if exists letters`)
	dbi.Exec(`create table letters(x varchar(1))`)
    dbi.Exec(`create procedure proc_w_resultset() begin insert into letters values('m'); insert into letters values('n'); select x from letters; select 1; select 2; insert into letters values('a'); end`)

	sql := `call proc_w_resultset`
	lists := make([]map[string]interface{},0)
	err = dbi.SelectProcContext(ctx, &lists, sql, nil)
	if err != nil {
		t.Errorf("Running select procedure failed %v", err)
	}
	if string(lists[0]["x"].(string)) != "m" {
		t.Errorf("%s m wanted", lists[0]["x"])
	}
	if string(lists[1]["x"].(string)) != "n" {
		t.Errorf("%s n wanted", lists[1]["x"])
	}

	dbi.Exec(`create procedure proc_w(IN x0 varchar(1),OUT y0 int) begin delete from letters; insert into letters values('m'); insert into letters values('n'); insert into letters values('p'); select x from letters where x=x0; insert into letters values('a'); set y0=100; end`)

	sql = `call proc_w`
	hash := make(map[string]interface{})
	lists = make([]map[string]interface{},0)
	err = dbi.SelectDoProcContext(ctx, &lists, hash, []interface{}{"y0"}, sql, nil, "m")
	if err != nil {
		t.Errorf("Running select do procedure failed %v", err)
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
	dbi.Exec(`drop procedure if exists proc_w`)
	dbi.Exec(`drop procedure if exists proc_w_resultset`)
	dbi.Exec(`drop table if exists letters`)
	db.Close()
}
