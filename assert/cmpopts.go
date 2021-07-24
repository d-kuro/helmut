package assert

import (
	"reflect"
	"strings"
)

// SortObjectsByNameField is the function used for cmpopts.SortSlices.
// If the struct slice has a "Name" field, compare and sort in lexicographic order.
func SortObjectsByNameField(x, y interface{}) bool {
	v1, ok := findNameField(x)
	if !ok {
		return false
	}

	v2, ok := findNameField(y)
	if !ok {
		return false
	}

	return v1 < v2
}

func findNameField(v interface{}) (string, bool) {
	rv := reflect.ValueOf(v)

	if rv.Kind() == reflect.Ptr {
		rv = reflect.ValueOf(v).Elem()
	}

	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		if strings.ToLower(rt.Field(i).Name) == "name" {
			if s, ok := rv.Field(i).Interface().(string); ok {
				return s, true
			}
		}
	}

	return "", false
}
