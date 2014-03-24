package jsonschema

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

type maximum struct {
	json.Number
	exclusive bool
}

func (m maximum) Validate(v interface{}) []ValidationError {
	normalized, err := normalizeNumber(v)
	if err != nil {
		return []ValidationError{ValidationError{err.Error()}}
	}
	var isLarger bool
	switch n := normalized.(type) {
	case int64:
		isLarger, err = m.isLargerThanInt(n)
	case float64:
		isLarger, err = m.isLargerThanFloat(n)
	default:
		return nil
	}
	if err != nil {
		return nil
	}
	if !isLarger {
		maxErr := fmt.Sprintf("Value must be smaller than %s.", m)
		return []ValidationError{ValidationError{maxErr}}
	}
	return nil
}

func (m maximum) isLargerThanInt(n int64) (bool, error) {
	if !strings.Contains(m.String(), ".") {
		max, err := m.Int64()
		if err != nil {
			return false, err
		}
		return max > n || !m.exclusive && max == n, nil
	} else {
		return m.isLargerThanFloat(float64(n))
	}
}

func (m maximum) isLargerThanFloat(n float64) (isLarger bool, err error) {
	max, err := m.Float64()
	if err != nil {
		return
	}
	return max > n || !m.exclusive && max == n, nil
}

func (m *maximum) SetSchema(v map[string]json.RawMessage) {
	value, ok := v["exclusiveMaximum"]
	if ok {
		// Ignore errors from Unmarshal. If exclusiveMaximum is a non boolean JSON
		// value we leave it as false.
		json.Unmarshal(value, &m.exclusive)
	}
	return
}

func (m *maximum) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &m.Number)
}

type minimum struct {
	json.Number
	exclusive bool
}

func (m minimum) Validate(v interface{}) []ValidationError {
	normalized, err := normalizeNumber(v)
	if err != nil {
		return []ValidationError{ValidationError{err.Error()}}
	}
	var isLarger bool
	switch n := normalized.(type) {
	case int64:
		isLarger, err = m.isLargerThanInt(n)
	case float64:
		isLarger, err = m.isLargerThanFloat(n)
	default:
		return nil
	}
	if err != nil {
		return nil
	}
	if isLarger {
		minErr := fmt.Sprintf("Value must be smaller than %s.", m)
		return []ValidationError{ValidationError{minErr}}
	}
	return nil
}

func (m minimum) isLargerThanInt(n int64) (bool, error) {
	if !strings.Contains(m.String(), ".") {
		min, err := m.Int64()
		if err != nil {
			return false, nil
		}
		return min > n || !m.exclusive && min == n, nil
	} else {
		return m.isLargerThanFloat(float64(n))
	}
}

func (m minimum) isLargerThanFloat(n float64) (isLarger bool, err error) {
	min, err := m.Float64()
	if err != nil {
		return
	}
	return min > n || !m.exclusive && min == n, nil
}

func (m *minimum) SetSchema(v map[string]json.RawMessage) {
	value, ok := v["exclusiveminimum"]
	if ok {
		// Ignore errors from Unmarshal. If exclusiveminimum is a non boolean JSON
		// value we leave it as false.
		json.Unmarshal(value, &m.exclusive)
	}
	return
}

func (m *minimum) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &m.Number)
}

type multipleOf int64

// Contrary to the spec, validation doesn't support floats in the schema
// or the data being validated. This is because of issues with math.Mod,
// e.g. math.Mod(0.0075, 0.0001) != 0.
func (m multipleOf) Validate(v interface{}) []ValidationError {
	normalized, err := normalizeNumber(v)
	if err != nil {
		return []ValidationError{ValidationError{err.Error()}}
	}
	n, ok := normalized.(int64)
	if !ok {
		return nil
	}
	if n%int64(m) != 0 {
		mulErr := ValidationError{fmt.Sprintf("Value must be a multiple of %d.", m)}
		return []ValidationError{mulErr}
	}
	return nil
}

func (m *multipleOf) UnmarshalJSON(b []byte) error {
	var n int64
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	*m = multipleOf(n)
	return nil
}

type maxLength int

func (m maxLength) Validate(v interface{}) []ValidationError {
	l, ok := v.(string)
	if !ok {
		return nil
	}
	if utf8.RuneCountInString(l) > int(m) {
		lenErr := ValidationError{fmt.Sprintf("String length must be shorter than %d characters.", m)}
		return []ValidationError{lenErr}
	}
	return nil
}

type minLength int

func (m minLength) Validate(v interface{}) []ValidationError {
	l, ok := v.(string)
	if !ok {
		return nil
	}
	if utf8.RuneCountInString(l) < int(m) {
		lenErr := ValidationError{fmt.Sprintf("String length must be shorter than %d characters.", m)}
		return []ValidationError{lenErr}
	}
	return nil
}

type pattern struct {
	regexp.Regexp
}

func (p pattern) Validate(v interface{}) []ValidationError {
	s, ok := v.(string)
	if !ok {
		return nil
	}
	if !p.MatchString(s) {
		patErr := ValidationError{fmt.Sprintf("String must match the pattern: \"%s\".", p.String())}
		return []ValidationError{patErr}
	}
	return nil
}

func (p *pattern) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	r, err := regexp.Compile(s)
	if err != nil {
		return err
	}
	p.Regexp = *r
	return nil
}

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

func (i *items) SetSchema(v map[string]json.RawMessage) {
	i.additionalAllowed = true
	value, ok := v["additionalItems"]
	if !ok {
		return
	}
	json.Unmarshal(value, &i.additionalAllowed)
	json.Unmarshal(value, &i.additionalItems)
	return
}

func (i *items) UnmarshalJSON(b []byte) error {
	if err1 := json.Unmarshal(b, &i.schema); err1 != nil {
		i.schema = nil
	}
	if err := json.Unmarshal(b, &i.schemaSlice); err != nil {
		i.schemaSlice = nil
		return err
	}
	return nil
}

type properties map[string]json.RawMessage

func (p properties) Validate(v interface{}) []ValidationError {
	var valErrs []ValidationError
	for schemaKey, schemaValue := range p {
		dataMap, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		if dataValue, ok := dataMap[schemaKey]; ok {
			var schema Schema
			err := json.Unmarshal(schemaValue, &schema)
			if err != nil {
				break
			}
			valErrs = append(valErrs, schema.Validate(dataValue)...)
		}
	}
	return valErrs
}
