package jsonschema

import (
	"encoding/json"
	"fmt"
)

type maxItems int

func (m maxItems) Validate(v interface{}) []ValidationError {
	l, ok := v.([]interface{})
	if !ok {
		return nil
	}
	if len(l) > int(m) {
		maxErr := ValidationError{fmt.Sprintf("Array must have fewer than %d items.", m)}
		return []ValidationError{maxErr}
	}
	return nil
}

type minItems int

func (m minItems) Validate(v interface{}) []ValidationError {
	l, ok := v.([]interface{})
	if !ok {
		return nil
	}
	if len(l) < int(m) {
		minErr := ValidationError{fmt.Sprintf("Array must have more than %d items.", m)}
		return []ValidationError{minErr}
	}
	return nil
}

// The spec[0] is useless for this keyword. The implemention here is based on the tests and this[1] guide.
//
// [0] http://json-schema.org/latest/json-schema-validation.html#anchor37
// [1] http://spacetelescope.github.io/understanding-json-schema/reference/array.html
type items struct {
	schema            *Schema
	schemaSlice       []*Schema
	additionalAllowed bool
	additionalItems   *Schema
}

func (i *items) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &i.schema); err == nil {
		return nil
	}
	i.schema = nil
	if err := json.Unmarshal(b, &i.schemaSlice); err != nil {
		return err
	}
	return nil
}

func (i *items) SetSchema(v map[string]json.RawMessage) error {
	i.additionalAllowed = true
	value, ok := v["additionalItems"]
	if !ok {
		return nil
	}
	json.Unmarshal(value, &i.additionalAllowed)
	json.Unmarshal(value, &i.additionalItems)
	return nil
}

func (i items) Validate(v interface{}) []ValidationError {
	var valErrs []ValidationError
	instances, ok := v.([]interface{})
	if !ok {
		return nil
	}
	if i.schema != nil {
		for _, value := range instances {
			valErrs = append(valErrs, i.schema.Validate(value)...)
		}
	} else if i.schemaSlice != nil {
		for pos, value := range instances {
			if pos <= len(i.schemaSlice)-1 {
				schema := i.schemaSlice[pos]
				valErrs = append(valErrs, schema.Validate(value)...)
			} else if i.additionalAllowed {
				if i.additionalItems == nil {
					continue
				}
				valErrs = append(valErrs, i.additionalItems.Validate(value)...)
			} else if !i.additionalAllowed {
				return []ValidationError{ValidationError{"Additional items aren't allowed."}}
			}
		}
	}
	return valErrs
}
