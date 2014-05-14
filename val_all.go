package jsonschema

import (
	"encoding/json"
	"fmt"
	"strings"
)

type allOf []Schema

func (a allOf) Validate(v interface{}) (valErrs []ValidationError) {
	for _, schema := range a {
		valErrs = append(valErrs, schema.Validate(v)...)
	}
	return
}

type anyOf []Schema

func (a anyOf) Validate(v interface{}) []ValidationError {
	for _, schema := range a {
		if schema.Validate(v) == nil {
			return nil
		}
	}
	return []ValidationError{
		ValidationError{"Validation failed for each schema in 'anyOf'."}}
}

type enum []interface{}

func (a enum) Validate(v interface{}) []ValidationError {
	for _, b := range a {
		if DeepEqual(v, b) {
			return nil
		}
	}
	return []ValidationError{
		ValidationError{fmt.Sprintf("Enum error. The data must be equal to one of these values %v.", a)}}
}

type not struct {
	Schema
}

func (n not) Validate(v interface{}) []ValidationError {
	schema := Schema{n.vals}
	if schema.Validate(v) == nil {
		return []ValidationError{ValidationError{"The 'not' schema didn't raise an error."}}
	}
	return nil
}

type oneOf []Schema

func (a oneOf) Validate(v interface{}) []ValidationError {
	var succeeded int
	for _, schema := range a {
		if schema.Validate(v) == nil {
			succeeded++
		}
	}
	if succeeded != 1 {
		return []ValidationError{ValidationError{
			fmt.Sprintf("Validation passed for %d schemas in 'oneOf'.", succeeded)}}
	}
	return nil
}

type typeValidator map[string]bool

func (t *typeValidator) UnmarshalJSON(b []byte) error {
	*t = make(typeValidator)
	var s string
	var l []string

	// The value of the "type" keyword can be a string or an array.
	if err := json.Unmarshal(b, &s); err != nil {
		err = json.Unmarshal(b, &l)
		if err != nil {
			return err
		}
	} else {
		l = []string{s}
	}

	for _, val := range l {
		(*t)[val] = true
	}
	return nil
}

func (t typeValidator) Validate(v interface{}) []ValidationError {
	var s string

	switch x := v.(type) {

	case string:
		s = "string"
	case bool:
		s = "boolean"
	case nil:
		s = "null"
	case []interface{}:
		s = "array"
	case map[string]interface{}:
		s = "object"

	case json.Number:
		if strings.Contains(x.String(), ".") {
			s = "number"
		} else {
			s = "integer"
		}
	case float64:
		s = "number"
	}

	_, ok := t[s]

	// The "number" type includes the "integer" type.
	if !ok && s == "integer" {
		_, ok = t["number"]
	}

	if !ok {
		types := make([]string, 0, len(t))
		for key := range t {
			types = append(types, key)
		}
		return []ValidationError{ValidationError{
			fmt.Sprintf("Value must be one of these types: %s.", types)}}
	}
	return nil
}
