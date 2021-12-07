package godbi

import (
	"encoding/json"
	"testing"
)

func TestInsert(t *testing.T) {
	str := `{
    "fks":["","adv_id","","adv_id","campaign_id_md5"],
    "table":"adv_campaign",
    "pks":["campaign_id"],
    "idAuto":"campaign_id",
	"musts":["adv_id","campaign_name"],
	"columns":["adv_id","campaign_name"]
	}`
	insert := new(Insert)
	err := json.Unmarshal([]byte(str), insert)
	if err != nil {
		t.Fatal(err)
	}
	if insert.Musts[1] != "campaign_name" ||
		insert.Columns[0] != "adv_id" {
		t.Errorf("%v", insert)
	}
}
