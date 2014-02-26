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
	switch n := v.(type) {
	case int64:
		isLarger = m.isLargerThanInt(n)
	case float64:
		isLarger = m.isLargerThanFloat(n)
	}
	if isLarger {
		minErr := ValidationError{fmt.Sprintf("Value must be larger than %s.", m)}
		return []ValidationError{minErr}
	}
	return nil
}

func (m minimum) isLargerThanInt(n int64) bool {
	if strings.Contains(m.String(), ".") {
		flt, err := m.Float64()
		if err != nil {
			return false
		}
		if flt > float64(n) {
			return true
		}
	} else {
		intg, err := m.Int64()
		if err != nil {
			return false
		}
		if intg > n {
			return true
		}
	}
	return false
}

func (m minimum) isLargerThanFloat(n float64) bool {
	flt, err := m.Float64()
	if err != nil {
		return false
	}
	if flt > n {
		return true
	}
	return false
}
