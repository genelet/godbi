package godbi

import (
	"testing"
)

func TestModel(t *testing.T) {
	model, err := NewModelJsonFile("model.json")
	if err != nil { t.Fatal(err) }
	t.Errorf("%#v", model)
	for k, v := range model.Actions {
		t.Errorf("%s=>%#v", k, v)
	}
/*
	if comp.CurrentIDAuto != "sid" || comp.Fks[1] != "cid" {
		t.Errorf("%#v", comp)
	}
	del := comp.Capability("delete")
	if del.Must[0] != "student_id" ||
		del.Extras[0].Name != "status" ||
		del.Extras[0].Pars[0] != "normal" {
		t.Errorf("%#v", del)
	}

	str, err := comp.ToHCL()
	if err != nil { t.Fatal(err) }
	comp1, err := ComponentFromHCL(str)
	if err != nil { t.Fatal(err) }
	if comp1.CurrentTable != comp.CurrentTable ||
		comp1.CurrentKeys[0] != comp.CurrentKeys[0] ||
		comp1.Fks[2] != comp.Fks[2] {
		t.Errorf("%#v", comp)
		t.Errorf("%#v", comp1)
	}
	if comp1.Capabilities[0].Nextpages[0].RelateItem["key1"] !=
		comp.Capabilities[0].Nextpages[0].RelateItem["key1"] {
		t.Errorf("%#v", comp.Capabilities[0].Nextpages[0])
		t.Errorf("%#v", comp1.Capabilities[0].Nextpages[0])
	}

	//bs, err := json.Marshal(comp)
	bs, err := json.MarshalIndent(comp, "", "  ")
	if err != nil { t.Fatal(err) }
	// t.Errorf("%s\n", bs)
	comp2 := new(Component)
	err = json.Unmarshal(bs, comp2)
	if err != nil { t.Fatal(err) }
	topics := comp.Capability("topics")
	rename := topics.Select.Rename
	if rename["x"][0] != "namex" || rename["x"][1] != "string" {
		t.Errorf("%#v\n", topics.Select)
	}
*/
}
