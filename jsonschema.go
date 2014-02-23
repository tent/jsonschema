package jsonschema

import (
	"encoding/json"
	"io"
)

func Parse(schemaBytes io.Reader) (*Schema, error) {
	var schema *Schema
	if err := json.NewDecoder(schemaBytes).Decode(&schema); err != nil {
		return nil, err
	}
	return schema, nil
}

func (s *Schema) Validate(dataStruct interface{}) []ValidationError {
	var valErrs []ValidationError
	typeString := typeSwitch(dataStruct)
	if s.Minimum != nil {
		var err error
		switch typeString {
		case "int64":
			err = IntMinimum(s, dataStruct.(int64))
		case "float64":
			err = FloatMinimum(s, dataStruct.(float64))
		default:
			err = nil
		}
		if err != nil {
			valErrs = append(valErrs, ValidationError{err.Error()})
		}
	}
	if s.Properties != nil && typeString == "map[string]interface{}" {
		for schemaKey, schemaValue := range *s.Properties {
			if dataValue, ok := dataStruct.(map[string]interface{})[schemaKey]; ok {
				valErrs = append(valErrs, schemaValue.Validate(dataValue)...)
			}
		}
	}
	return valErrs
}

type Schema struct {
	Minimum    *json.Number       `json:"minimum"`
	Properties *map[string]Schema `json:"properties"`
}

type ValidationError struct {
	Description string
}
