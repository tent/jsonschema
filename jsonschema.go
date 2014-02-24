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
	data, typeString := normalizeType(dataStruct)
	if s.Minimum != nil {
		err := Minimum(s, data)
		if err != nil {
			valErrs = append(valErrs, ValidationError{err.Error()})
		}
	}
	if s.Properties != nil && typeString == "map[string]interface{}" {
		for schemaKey, schemaValue := range *s.Properties {
			if dataValue, ok := data.(map[string]interface{})[schemaKey]; ok {
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
