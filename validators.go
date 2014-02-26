package jsonschema

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type properties map[string]json.RawMessage

func (p properties) Validate(v interface{}) ([]ValidationError, error) {
	var valErrs []ValidationError
	for schemaKey, schemaValue := range p {
		dataMap, ok := v.(map[string]interface{})
		if !ok {
			return valErrs, errors.New("Properties must be of the type `map[string]interface{}`.")
		}
		if dataValue, ok := dataMap[schemaKey]; ok {
			var schema Schema
			err := json.Unmarshal(schemaValue, &schema)
			if err != nil {
				break
			}
			newErrors, err := schema.Validate(dataValue)
			if err != nil {
				return valErrs, err
			}
			valErrs = append(valErrs, newErrors...)
		}
	}
	return valErrs, nil
}

type minimum struct {
	json.Number
}

func (m minimum) Validate(unnormalized interface{}) ([]ValidationError, error) {
	v, err := normalizeNumber(unnormalized)
	if err != nil {
		return []ValidationError{}, err
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
		return []ValidationError{minErr}, nil
	}
	return []ValidationError{}, nil
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
