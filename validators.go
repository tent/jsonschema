package jsonschema

import (
	"encoding/json"
	"fmt"
	"strings"
)

func Properties(val interface{}) func(interface{}) []ValidationError {
	props, ok := val.(map[string]interface{})
	if !ok {
		return func(dataStruct interface{}) []ValidationError {
			return []ValidationError{}
		}
	}
	return func(dataStruct interface{}) []ValidationError {
		var valErrs []ValidationError
		for schemaKey, schemaValue := range props {
			dataMap, ok := dataStruct.(map[string]interface{})
			if !ok {
				return valErrs
			}
			if dataValue, ok := dataMap[schemaKey]; ok {
				var schema Schema
				bts, err := json.Marshal(schemaValue)
				if err != nil {
					break
				}
				err = json.Unmarshal(bts, &schema)
				if err != nil {
					break
				}
				valErrs = append(valErrs, schema.Validate(dataValue)...)
			}
		}
		return valErrs
	}
}

func Minimum(val interface{}) func(interface{}) []ValidationError {
	min, ok := val.(json.Number)
	if !ok {
		return func(dataStruct interface{}) []ValidationError {
			return []ValidationError{}
		}
	}
	return func(dataStruct interface{}) []ValidationError {
		var isLarger bool
		switch dataStruct.(type) {
		case int64:
			isLarger = isLargerThanInt(min, dataStruct.(int64))
		case float64:
			isLarger = isLargerThanFloat(min, dataStruct.(float64))
		}
		if isLarger {
			minErr := ValidationError{fmt.Sprintf("Value must be larger than %s.", min)}
			return []ValidationError{minErr}
		}
		return nil
	}
}

func isLargerThanInt(min json.Number, data int64) bool {
	if strings.Contains(min.String(), ".") {
		flt, err := min.Float64()
		if err != nil {
			return false
		}
		if flt > float64(data) {
			return true
		}
	} else {
		intg, err := min.Int64()
		if err != nil {
			return false
		}
		if intg > data {
			return true
		}
	}
	return false
}

func isLargerThanFloat(min json.Number, data float64) bool {
	flt, err := min.Float64()
	if err != nil {
		return false
	}
	if flt > data {
		return true
	}
	return false
}
