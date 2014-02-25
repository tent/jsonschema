package jsonschema

import (
	"bytes"
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
	if min, ok := schemaMap["minimum"]; ok {
		s.Validators = append(s.Validators, Minimum(min))
	}
	if prop, ok := schemaMap["properties"]; ok {
		s.Validators = append(s.Validators, Properties(prop))
	}
	return nil
}

type Schema struct {
	Validators []func(interface{}) ([]ValidationError, error)
}

type ValidationError struct {
	Description string
}
