package godbi

import (
	"strings"
)

func hasValue(extra interface{}) bool {
	if extra == nil {
		return false
	}
	switch v := extra.(type) {
	case []string:
		if len(v) == 0 {
			return false
		}
	case []interface{}:
		if len(v) == 0 {
			return false
		}
	case []*Joint:
		if len(v) == 0 {
			return false
		}
	case map[string]string:
		if len(v) == 0 {
			return false
		}
	case map[string]interface{}:
		if len(v) == 0 {
			return false
		}
	case []map[string]interface{}:
		if len(v) == 0 {
			return false
		}
	default:
	}
	return true
}

func stripchars(str, chr string) string {
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

func grep(vs []string, t string) bool {
	return index(vs, t) >= 0
}
