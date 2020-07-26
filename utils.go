package godbi

import (
	"fmt"
	"net/url"
	"strings"
)

func HasValue(extra interface{}) bool {
	if extra == nil {
		return false
	}
	switch v := extra.(type) {
	case []string:
		if len(v) == 0 { return false }
	case []interface{}:
		if len(v) == 0 { return false }
	case []*Table:
		if len(v) == 0 { return false }
	case url.Values:
		if len(v) == 0 { return false }
	case []url.Values:
		if len(v) == 0 { return false }
	case map[string]string:
		if len(v) == 0 { return false }
	case map[string]interface{}:
		if len(v) == 0 { return false }
	case []map[string]interface{}:
		if len(v) == 0 { return false }
	default:
	}
	return true
}

func Interface2String(v interface{}) string {
	switch u := v.(type) {
	case []uint8:
		return string(u)
	case int8, uint8, int, uint, int32, uint32, int64, uint64:
		return fmt.Sprintf("%d", u)
	case float32, float64:
		return fmt.Sprintf("%f", u)
	case string:
		return u
	default:
		return v.(string)
	}
}

func stripchars(chr, str string) string {
	return strings.Map(func(r rune) rune {
		if strings.IndexRune(chr, r) < 0 {
			return r
		}
		return -1
	}, str)
}

func filtering(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func mapping(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

func index(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

func Grep(vs []string, t string) bool {
	return index(vs, t) >= 0
}
