package jsonschema

import (
	"encoding/json"
	"fmt"
	"strings"
)

type properties map[string]json.RawMessage

func (p properties) Validate(v interface{}) []ValidationError {
	var valErrs []ValidationError
	for schemaKey, schemaValue := range p {
		dataMap, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		if dataValue, ok := dataMap[schemaKey]; ok {
			var schema Schema
			err := json.Unmarshal(schemaValue, &schema)
			if err != nil {
				break
			}
			valErrs = append(valErrs, schema.Validate(dataValue)...)
		}
	}
	return valErrs
}

type minimum struct {
	json.Number
}

func (m minimum) Validate(unnormalized interface{}) []ValidationError {
	v, err := normalizeNumber(unnormalized)
	if err != nil {
		return []ValidationError{ValidationError{err.Error()}}
	}
	var isLarger bool
	switch v.(type) {
	case int64:
		isLarger = m.isLargerThanInt(v.(int64))
	case float64:
		isLarger = m.isLargerThanFloat(v.(float64))
	}
	if isLarger {
		minErr := ValidationError{fmt.Sprintf("Value must be larger than %s.", m)}
		return []ValidationError{minErr}
	}
	return []ValidationError{}
}

func (min minimum) isLargerThanInt(data int64) bool {
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

func (min minimum) isLargerThanFloat(data float64) bool {
	flt, err := min.Float64()
	if err != nil {
		return false
	}
	if flt > data {
		return true
	}
	return false
}
