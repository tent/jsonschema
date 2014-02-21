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
	var validationErrors []ValidationError
	if s.Minimum != nil {
		if !Minimum(s, dataStruct) {
			minimumError := ValidationError{"Value below minimum."}
			validationErrors = append(validationErrors, minimumError)
		}
	}
	return validationErrors
}

type Schema struct {
	Minimum *float64 `json:"minimum"`
}

type ValidationError struct {
	Description string
}
