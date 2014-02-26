package jsonschema

import (
	"encoding/json"
	"io"
	"reflect"
)

var validatorMap = map[string]reflect.Type{
	"minimum":    reflect.TypeOf(minimum{}),
	"properties": reflect.TypeOf(properties{})}

type Validator interface {
	Validate(interface{}) ([]ValidationError, error)
}

func Parse(schemaBytes io.Reader) (*Schema, error) {
	var schema *Schema
	if err := json.NewDecoder(schemaBytes).Decode(&schema); err != nil {
		return nil, err
	}
	return schema, nil
}

func (s *Schema) Validate(v interface{}) ([]ValidationError, error) {
	var valErrs []ValidationError
	for _, validator := range s.Vals {
		newErrors, err := validator.Validate(v)
		if err != nil {
			return valErrs, err
		}
		valErrs = append(valErrs, newErrors...)
	}
	return valErrs, nil
}

func (s *Schema) UnmarshalJSON(bts []byte) error {
	schemaMap := make(map[string]json.RawMessage)
	if err := json.Unmarshal(bts, &schemaMap); err != nil {
		return err
	}
	for schemaKey, schemaValue := range schemaMap {
		if typ, ok := validatorMap[schemaKey]; ok {
			var newValidator = reflect.New(typ).Interface().(Validator)
			if err := json.Unmarshal(schemaValue, newValidator); err != nil {
				return err
			}
			s.Vals = append(s.Vals, newValidator)
		}
	}
	return nil
}

type Schema struct {
	Vals []Validator
}

type ValidationError struct {
	Description string
}
