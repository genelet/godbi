package godbi

import (
	"encoding/json"
	"testing"
)

func TestTable(t *testing.T) {
	str := `{
    "fks":["","adv_id","","adv_id","campaign_id_md5"],
    "table":"adv_campaign",
    "pks":["campaign_id"],
    "id_auto":"campaign_id"
	}`
	table := new(Table)
	err := json.Unmarshal([]byte(str), table)
	if err != nil {
		t.Fatal(err)
	}
	if table.Fks[4] != "campaign_id_md5" ||
		table.CurrentTable != "adv_campaign" {
		t.Errorf("%v", table)
	}
}
