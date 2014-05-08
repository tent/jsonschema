package jsonschema

import (
	"bytes"
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
	"format":    reflect.TypeOf(format("")),

	// Arrays
	"maxItems": reflect.TypeOf(maxItems(0)),
	"minItems": reflect.TypeOf(minItems(0)),
	"items":    reflect.TypeOf(items{}),

	// Objects
	"dependencies":      reflect.TypeOf(dependencies{}),
	"maxProperties":     reflect.TypeOf(maxProperties(0)),
	"minProperties":     reflect.TypeOf(minProperties(0)),
	"patternProperties": reflect.TypeOf(patternProperties{}),
	"properties":        reflect.TypeOf(properties{}),
	"required":          reflect.TypeOf(required{}),

	// All types
	"allOf":       reflect.TypeOf(allOf{}),
	"anyOf":       reflect.TypeOf(anyOf{}),
	"definitions": reflect.TypeOf(definitions{}),
	"enum":        reflect.TypeOf(enum{}),
	"not":         reflect.TypeOf(not{}),
	"oneOf":       reflect.TypeOf(oneOf{}),
	"$ref":        reflect.TypeOf(ref("")),
	"type":        reflect.TypeOf(typeValidator{})}

type Validator interface {
	Validate(interface{}) []ValidationError
}

func Parse(schemaBytes io.Reader) (*Schema, error) {
	var schema *Schema
	if err := json.NewDecoder(schemaBytes).Decode(&schema); err != nil {
		return nil, err
	}
	schema.resolveRefs()
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
	s.vals = make(map[string]Validator, len(schemaMap))
	for schemaKey, schemaValue := range schemaMap {

		var newValidator Validator
		if typ, ok := validatorMap[schemaKey]; ok {
			newValidator = reflect.New(typ).Interface().(Validator)
		} else {
			// Even if we don't recognize a schema key, we unmarshal its contents anyway
			// because it might contain embedded schemas referenced elsewhere in the document.
			newValidator = new(other)
		}

		decoder := json.NewDecoder(bytes.NewReader(schemaValue))
		decoder.UseNumber()
		if err := decoder.Decode(newValidator); err != nil {
			continue
		}
		// Make changes to a validator based on its neighbors, if appropriate.
		if v, ok := newValidator.(SchemaSetter); ok {
			v.SetSchema(schemaMap)
		}
		s.vals[schemaKey] = newValidator
	}
	return nil
}

// A SchemaSetter is a validator (such as maximum) whose validate method depends
// on neighboring schema keys (such as exclusiveMaximum). When a SchemaSetter is
// unmarshaled from JSON, SetSchema is called on its neighbors to see if any of
// them are relevant to the validator being unmarshaled.
type SchemaSetter interface {
	SetSchema(map[string]json.RawMessage) error
}

type Schema struct {
	vals     map[string]Validator
	resolved bool
}

type ValidationError struct {
	Description string
}
