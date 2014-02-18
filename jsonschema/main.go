package jsonschema

import (
	"io"
)

func Parse(schemaBytes io.Reader) (*Schema, error) {
	var emptySchema *Schema
	return emptySchema, nil
}

func (s *Schema) Validate(dataStruct interface{}) []ValidationError {
	notImplemented := ValidationError{"Validation not implemented yet."}
	return []ValidationError{notImplemented}
}

type Schema struct {
}

type ValidationError struct {
	Description string
}
