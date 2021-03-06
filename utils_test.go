package godbi

import (
	"testing"
)

func TestUtils(t *testing.T) {
	str := "abcdefg_+hijk=="
	newStr := stripchars("df+=", str)
	if "abceg_hijk" != newStr {
		t.Errorf("%s %s wanted", str, newStr)
	}
	x := []string{str, newStr, "abc"}
	if grep(x, "abcZ") {
		t.Errorf("%s wrong matched", "abcZ")
	}
	if grep(x, "abc") == false {
		t.Errorf("%s matched", "abc")
	}
	if grep([]string{"child","tid"}, "tid")==false {
		t.Errorf("%#v does not match %s", []string{"child","tid"}, "tid")
	}
	x1 := []interface{}{"a","b"}
	x2 := map[string]interface{}{"a":"b"}
	x3 := make([]interface{},0)
	x4 := make(map[string]interface{})
	if !hasValue(x1) {
		t.Errorf("%v", x1)
	}
	if !hasValue(x2) {
		t.Errorf("%v", x2)
	}
	if hasValue(x3) {
		t.Errorf("%v", x3)
	}
	if hasValue(x4) {
		t.Errorf("%v", x4)
	}

	bs := []byte("hello")
	x4["a1"] = []uint8{bs[0],bs[1],bs[2],bs[3],bs[4]}
	x4["a2"] = uint8(1)
	x4["a3"] = int(2)
	x4["a4"] = int64(3)
	x4["a5"] = float32(4.5)
	x4["a6"] = float32(6.7)
	x4["a7"] = "world"
	if interface2String(x4["a1"]) != "hello" {
		t.Errorf("%v", interface2String(x4["a1"]))
	}
	if interface2String(x4["a2"]) != "1" {
		t.Errorf("%v", interface2String(x4["a2"]))
	}
	if interface2String(x4["a3"]) != "2" {
		t.Errorf("%v", interface2String(x4["a3"]))
	}
	y := interface2String(x4["a6"])
	if y[:3] != "6.7" {
		t.Errorf("%v", y)
	}
}
