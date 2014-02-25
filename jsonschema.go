package jsonschema

import (
	"bytes"
	"encoding/json"
	"io"
)

var validatorMap = map[string]func(json.RawMessage) func(interface{}) ([]ValidationError, error){
	"minimum":    Minimum,
	"properties": Properties}

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
	for _, validator := range s.Validators {
		newErrors, err := validator(data)
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
		if validatorCreator, ok := validatorMap[schemaKey]; ok {
			s.Validators = append(s.Validators, validatorCreator(schemaValue))
		}
	}
	return nil
}

type Schema struct {
	Validators []func(interface{}) ([]ValidationError, error)
}

type ValidationError struct {
	Description string
}
