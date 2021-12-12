package godbi

import (
	"encoding/json"
	"testing"
)

func TestTable(t *testing.T) {
	str := `{
    "fks":[{"fkColumn":"adv_id","column":"adv_id"}],
    "tableName":"adv_campaign",
    "pks":["campaign_id"],
    "idAuto":"campaign_id",
	"columns":[{
		"columnName":"adv_id",
		"label":"adv_id",
		"typeName":"int",
		"notnull":true,
		"auto":false
		},{
		"columnName":"campaign_name",
		"label":"campaign_name",
		"typeName":"string",
		"notnull":true
		},{
		"columnName":"campaign_id",
		"label":"campaign_id",
		"typeName":"int",
		"notnull":true,
		"auto":true
		}
	]
	}`
	table := new(Table)
	err := json.Unmarshal([]byte(str), table)
	if err != nil {
		t.Fatal(err)
	}
	if table.TableName != "adv_campaign" {
		t.Errorf("%v", table)
	}
	inCols := table.insertCols()
	if inCols[0] != "adv_id" || inCols[1] != "campaign_name" {
		t.Errorf("%v", table)
	}
}
