package jsonschema

import (
	"encoding/json"
	"io"
	"reflect"
)

var validatorMap = map[string]reflect.Type{
	// Numbers
	"minimum": reflect.TypeOf(minimum{}),

	// Strings
	"maxLength": reflect.TypeOf(maxLength{}),
	"minLength": reflect.TypeOf(minLength{}),
	"pattern":   reflect.TypeOf(pattern{}),

	// Objects
	"properties": reflect.TypeOf(properties{})}

type Validator interface {
	Validate(interface{}) []ValidationError
}

func Parse(schemaBytes io.Reader) (*Schema, error) {
	var schema *Schema
	if err := json.NewDecoder(schemaBytes).Decode(&schema); err != nil {
		return nil, err
	}
	return schema, nil
}

func (s *Schema) Validate(v interface{}) []ValidationError {
	var valErrs []ValidationError
	for _, validator := range s.vals {
		valErrs = append(valErrs, validator.Validate(v)...)
	}
	return valErrs
}

func (s *Schema) UnmarshalJSON(bts []byte) error {
	schemaMap := make(map[string]json.RawMessage)
	if err := json.Unmarshal(bts, &schemaMap); err != nil {
		return err
	}
	s.vals = make([]Validator, 0, len(schemaMap))
	for schemaKey, schemaValue := range schemaMap {
		if typ, ok := validatorMap[schemaKey]; ok {
			var newValidator = reflect.New(typ).Interface().(Validator)
			if err := json.Unmarshal(schemaValue, newValidator); err != nil {
				continue
			}
			if v, ok := newValidator.(SchemaSetter); ok {
				v.SetSchema(schemaMap)
			}
			s.vals = append(s.vals, newValidator)
		}
	}
	return nil
}

type SchemaSetter interface {
	SetSchema(map[string]json.RawMessage)
}

type Schema struct {
	vals []Validator
}

type ValidationError struct {
	Description string
}
