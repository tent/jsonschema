package jsonschema

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type properties map[string]json.RawMessage

func (p properties) Validate(dataStruct interface{}) ([]ValidationError, error) {
	var valErrs []ValidationError
	for schemaKey, schemaValue := range p {
		dataMap, ok := dataStruct.(map[string]interface{})
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
	Val json.Number
}

func (m *minimum) UnmarshalJSON(bts []byte) error {
	var whats json.Number
	decoder := json.NewDecoder(bytes.NewReader(bts))
	decoder.UseNumber()
	if err := decoder.Decode(&whats); err != nil {
		return err
	}
	m.Val = whats
	return nil
}

func (m minimum) Validate(dataStruct interface{}) ([]ValidationError, error) {
	var isLarger bool
	switch dataStruct.(type) {
	case int64:
		isLarger = isLargerThanInt(m.Val, dataStruct.(int64))
	case float64:
		isLarger = isLargerThanFloat(m.Val, dataStruct.(float64))
	}
	if isLarger {
		minErr := ValidationError{fmt.Sprintf("Value must be larger than %s.", m.Val)}
		return []ValidationError{minErr}, nil
	}
	return []ValidationError{}, nil
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
