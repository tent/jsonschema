package jsonschema

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
)

var validatorMap = map[string]Validator{
	"minimum":    minimum{},
	"properties": properties{}}

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

func (s *Schema) Validate(dataStruct interface{}) ([]ValidationError, error) {
	data, err := normalizeType(dataStruct)
	var valErrs []ValidationError
	if err != nil {
		return valErrs, err
	}
	for _, validator := range s.Vals {
		newErrors, err := validator.Validate(data)
		if err != nil {
			return valErrs, err
		}
		valErrs = append(valErrs, newErrors...)
	}
	return valErrs, nil
}

func (s *Schema) UnmarshalJSON(bts []byte) error {
	var schemaMap map[string]json.RawMessage
	if err := json.NewDecoder(bytes.NewReader(bts)).Decode(&schemaMap); err != nil {
		return err
	}
	for schemaKey, schemaValue := range schemaMap {
		if exampleOfType, ok := validatorMap[schemaKey]; ok {
			var newVal = reflect.New(reflect.TypeOf(exampleOfType)).Interface()
			if err := json.Unmarshal(schemaValue, newVal); err != nil {
				return err
			}
			s.Vals = append(s.Vals, newVal.(Validator))
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
