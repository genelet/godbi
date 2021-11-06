package godbi

import (
	"testing"
)

func TestUtils(t *testing.T) {
	str := "abcdefg_+hijk=="
	newStr := stripchars(str, "df+=")
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
	if grep([]string{"child", "tid"}, "tid") == false {
		t.Errorf("%#v does not match %s", []string{"child", "tid"}, "tid")
	}
	x1 := []interface{}{"a", "b"}
	x2 := map[string]interface{}{"a": "b"}
	x3 := make([]interface{}, 0)
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
}
