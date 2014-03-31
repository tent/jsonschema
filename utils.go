package jsonschema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// normalizeNumber accepts any input and, if it is a supported number type,
// converts it to either int64 or float64. normalizeNumber raises an error
// if the input is an explicitly unsupported number type.
func normalizeNumber(v interface{}) (n interface{}, err error) {
	switch t := v.(type) {

	case json.Number:
		if strings.Contains(t.String(), ".") {
			n, err = t.Float64()
		} else {
			n, err = t.Int64()
		}

	case float32:
		n = float64(t)
	case float64:
		n = t

	case int:
		n = int64(t)
	case int8:
		n = int64(t)
	case int16:
		n = int64(t)
	case int32:
		n = int64(t)
	case int64:
		n = t

	case uint8:
		n = int64(t)
	case uint16:
		n = int64(t)
	case uint32:
		n = int64(t)
	case uint64:
		n = t
		err = fmt.Errorf("%s is not a supported type.", reflect.TypeOf(v))

	default:
		n = t
	}

	return
}

// Compare a data instance to a schema instance.
//
// Schema instances are always json.Number, never int64 or float64
// so we don't have to deal with the latter two types.
func isEqual(v, a interface{}) bool {
	switch b := a.(type) {
	case string:
		val, ok := v.(string)
		if ok {
			return string(val) == b
		}
	case bool:
		val, ok := v.(bool)
		if ok {
			return bool(val) == b
		}
	case nil:
		return v == nil
	case []interface{}:
		val, ok := v.([]interface{})
		if ok {
			if len(val) == len(b) {
				for key := range b {
					if !isEqual(val[key], b[key]) {
						return false
					}
				}
				return true
			}
		}
	case map[string]interface{}:
		val, ok := v.(map[string]interface{})
		if ok {
			if len(val) == len(b) {
				for key := range b {
					if !isEqual(val[key], b[key]) {
						return false
					}
				}
				return true
			}
		}
	// The reason this entire function can't be replaced by reflect.DeepEqual
	// is that DeepEqual doesn't know to compare json.Number to int64/float64.
	//
	// This wouldn't be a problem if json.Number could only occur at the top level
	// of the schema data because we could check for it before calling DeepEqual,
	// but json.Number can also be embedded in a slice or a map.
	case json.Number:
		z, err := normalizeNumber(v)
		if err != nil {
			return false
		}
		i, ok := z.(int64)
		if ok {
			c, err := b.Int64()
			if err != nil {
				return false
			}
			return i == c
		}
		f, ok := z.(float64)
		if ok {
			c, err := b.Float64()
			if err != nil {
				return false
			}
			return f == c
		}
	}
	return false
}
