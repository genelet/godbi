package godbi

import (
	"encoding/json"
	"testing"
)

func TestConnection(t *testing.T) {
	str := `{"model":"adv_campaign", "action":"edit", "relateExtra":{"campaign_id":"c_id"}, "relateArgs":{"x":"firstname"}}`
	page := new(Connection)
	err := json.Unmarshal([]byte(str), page)
	if err != nil {
		t.Fatal(err)
	}

	item := map[string]interface{}{"x": "a", "campaign_id": 123}
	nextArgs := page.NextArgs(item)
	if nextArgs["firstname"] != "a" {
		t.Errorf("%#v", nextArgs)
	}
	nextExtra := page.NextExtra(item)
	if nextExtra["c_id"] != 123 {
		t.Errorf("%#v", nextExtra)
	}

	extra := map[string]interface{}{"y": "b", "asset": "what"}
	hash := MergeExtra(nextExtra, extra)
	if hash["y"].(string) != "b" ||
		hash["asset"].(string) != "what" ||
		hash["c_id"].(int) != 123 {
		t.Errorf("%#v", hash)
	}

	item = map[string]interface{}{"x": "a", "item_id": 123}
	arg := map[string]interface{}{"y": "b", "asset": "what"}
	cArg  := CloneArgs(arg).(map[string]interface{})
	aArg  := MergeArgs(arg, item).(map[string]interface{})
	if len(cArg)!=2 || cArg["y"]!="b" || cArg["asset"]!="what" {
		t.Errorf("%#v", cArg)
	}
	if len(aArg)!=4 || aArg["item_id"]!=123 {
		t.Errorf("%#v", aArg)
	}

	args:= []map[string]interface{}{{"y": "b", "asset": "what"},
{"y": "bb", "asset": "whatwhat", "size_id":777}}
	cArgs := CloneArgs(args).([]map[string]interface{})
	aArgs := MergeArgs(args, item).([]map[string]interface{})
    //[]map[string]interface {}{map[string]interface {}{"asset":"what", "y":"b"}, map[string]interface {}{"asset":"whatwhat", "size_id":777, "y":"bb"}}
	if len(cArgs)!=2 || cArgs[0]["y"]!="b" || cArgs[1]["asset"]!="whatwhat" {
		t.Errorf("%#v", cArgs)
	}
    //[]map[string]interface {}{map[string]interface {}{"asset":"what", "item_id":123, "x":"a", "y":"b"}, map[string]interface {}{"asset":"whatwhat", "item_id":123, "size_id":777, "x":"a", "y":"bb"}}
	if len(aArgs)!=2 || len(aArgs[0])!=4 || aArgs[0]["item_id"]!=123 || len(aArgs[1])!=5 || aArgs[1]["item_id"]!=123 || aArgs[1]["size_id"]!=777 {
		t.Errorf("%#v", aArgs)
	}
}
