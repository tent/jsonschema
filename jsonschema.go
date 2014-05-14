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
	"additionalItems": reflect.TypeOf(additionalItems{}),
	"maxItems":        reflect.TypeOf(maxItems(0)),
	"minItems":        reflect.TypeOf(minItems(0)),
	"items":           reflect.TypeOf(items{}),

	// Objects
	"additionalProperties": reflect.TypeOf(additionalProperties{}),
	"dependencies":         reflect.TypeOf(dependencies{}),
	"maxProperties":        reflect.TypeOf(maxProperties(0)),
	"minProperties":        reflect.TypeOf(minProperties(0)),
	"patternProperties":    reflect.TypeOf(patternProperties{}),
	"properties":           reflect.TypeOf(properties{}),
	"required":             reflect.TypeOf(required{}),

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
	for _, node := range s.nodes {
		valErrs = append(valErrs, node.Validator.Validate(v)...)
	}
	return valErrs
}

func (s *Schema) UnmarshalJSON(bts []byte) error {
	schemaMap := make(map[string]json.RawMessage)
	if err := json.Unmarshal(bts, &schemaMap); err != nil {
		return err
	}
	s.nodes = make(map[string]*Node, len(schemaMap))
	for schemaKey, schemaValue := range schemaMap {

		var newNode Node
		newNode.EmbeddedSchemas = make(map[string]*Schema)

		if typ, ok := validatorMap[schemaKey]; ok {
			newNode.Validator = reflect.New(typ).Interface().(Validator)
		} else {
			// Even if we don't recognize a schema key, we unmarshal its contents anyway
			// because it might contain embedded schemas referenced elsewhere in the document.
			newNode.Validator = new(other)
		}

		if v, ok := newNode.Validator.(SchemaEmbedder); ok {
			v.LinkEmbedded(newNode.EmbeddedSchemas)
		}
		decoder := json.NewDecoder(bytes.NewReader(schemaValue))
		decoder.UseNumber()
		if err := decoder.Decode(newNode.Validator); err != nil {
			continue
		}

		// Make changes to a validator based on its neighbors, if appropriate.
		//
		// NOTE: this is deprecated in favor of UsesNeighboringSchemas.
		if v, ok := newNode.Validator.(SchemaSetter); ok {
			v.SetSchema(schemaMap)
		}

		s.nodes[schemaKey] = &newNode
	}
	for _, n := range s.nodes {
		if v, ok := n.Validator.(UsesNeighboringSchemas); ok {
			v.GetNeighboringSchemas(s.nodes)
		}
	}
	return nil
}

type UsesNeighboringSchemas interface {
	GetNeighboringSchemas(map[string]*Node)
}

// A SchemaSetter is a validator (such as maximum) whose validate method depends
// on neighboring schema keys (such as exclusiveMaximum). When a SchemaSetter is
// unmarshaled from JSON, SetSchema is called on its neighbors to see if any of
// them are relevant to the validator being unmarshaled.
//
// NOTE: this is deprecated in favor of UsesNeighboringSchemas.
type SchemaSetter interface {
	SetSchema(map[string]json.RawMessage) error
}

type Schema struct {
	nodes    map[string]*Node
	resolved bool
}

type Node struct {
	EmbeddedSchemas map[string]*Schema
	Validator       Validator
}

type ValidationError struct {
	Description string
}
