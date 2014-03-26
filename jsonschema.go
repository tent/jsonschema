package jsonschema

import (
	"encoding/json"
	"io"
	"reflect"
)

var validatorMap = map[string]reflect.Type{
	// Numbers
	"maximum":    reflect.TypeOf(maximum{}),
	"minimum":    reflect.TypeOf(minimum{}),
	"multipleOf": reflect.TypeOf(multipleOf(0)),

	// Strings
	"maxLength": reflect.TypeOf(maxLength(0)),
	"minLength": reflect.TypeOf(minLength(0)),
	"pattern":   reflect.TypeOf(pattern{}),

	// Arrays
	"maxItems": reflect.TypeOf(maxItems(0)),
	"minItems": reflect.TypeOf(minItems(0)),
	"items":    reflect.TypeOf(items{}),

	// Objects
	"properties": reflect.TypeOf(properties{}),

	// All types
	"allOf": reflect.TypeOf(allOf{}),
	"anyOf": reflect.TypeOf(anyOf{}),
	"not":   reflect.TypeOf(not{}),
	"oneOf": reflect.TypeOf(oneOf{}),
	"type":  reflect.TypeOf(typeValidator{})}

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
