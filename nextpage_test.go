package godbi

import (
	"encoding/json"
	"testing"
)

func TestNextpage(t *testing.T) {
	str := `{"model":"adv_campaign", "action":"edit", "relateItem":{"campaign_id":"c_id"}}`
	page := new(Nextpage)
	err := json.Unmarshal([]byte(str), page)
	if err != nil {
		t.Fatal(err)
	}

	item := map[string]interface{}{"x": "a", "campaign_id": 123}
	extra := map[string]interface{}{"y": "b", "asset": "what"}
	hashi, ok := page.refresh(item, extra)
	// hash has key "c_id"
	hash := hashi.(map[string]interface{})
	if !ok ||
		hash["asset"].(string) != "what" ||
		hash["c_id"].(int) != 123 {
		t.Errorf("%t, %#v", ok, hash)
	}

	item = map[string]interface{}{"x": "a", "item_id": 123}
	extra = map[string]interface{}{"y": "b", "asset": "what"}
	hashi, ok = page.refresh(item, extra)
	if ok {
		t.Errorf("%t, %#v", ok, hashi)
	}
	hash = hashi.(map[string]interface{})
	if hash["asset"].(string) != "what" ||
		hash["c_id"] != nil {
		t.Errorf("%t, %#v", ok, hash)
	}
}
