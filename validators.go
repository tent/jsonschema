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
	exclusive bool
}

func (m minimum) Validate(unnormalized interface{}) []ValidationError {
	v, err := normalizeNumber(unnormalized)
	if err != nil {
		return []ValidationError{ValidationError{err.Error()}}
	}
	var isLarger int
	switch n := v.(type) {
	case int64:
		isLarger = m.isLargerThanInt(n)
	case float64:
		isLarger = m.isLargerThanFloat(n)
	default:
		return nil
	}
	if isLarger > 0 || (m.exclusive && isLarger == 0) {
		minErr := ValidationError{fmt.Sprintf("Value must be larger than %s.", m)}
		return []ValidationError{minErr}
	}
	return nil
}

func (m *minimum) SetSchema(v map[string]json.RawMessage) {
	value, ok := v["exclusiveMinimum"]
	if ok {
		// Ignore errors from Unmarshal. If exclusiveMinimum is a non boolean JSON
		// value we leave it as false.
		json.Unmarshal(value, &m.exclusive)
	}
}

func (m *minimum) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &m.Number)
}

func (m minimum) isLargerThanInt(n int64) int {
	if !strings.Contains(m.String(), ".") {
		intg, err := m.Int64()
		if err != nil {
			return 0
		}
		if intg > n {
			return 1
		} else if intg < n {
			return -1
		}
	} else {
		flt, err := m.Float64()
		if err != nil {
			return 0
		}
		if flt > float64(n) {
			return 1
		} else if flt < float64(n) {
			return -1
		}
	}
	return 0
}

func (m minimum) isLargerThanFloat(n float64) int {
	flt, err := m.Float64()
	if err != nil {
		return 0
	}
	if flt > n {
		return 1
	} else if flt < n {
		return -1
	}
	return 0
}
