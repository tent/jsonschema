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

func (m *maximum) SetSchema(v map[string]json.RawMessage) error {
	value, ok := v["exclusiveMaximum"]
	if ok {
		// Ignore errors from Unmarshal. If exclusiveMaximum is a non boolean JSON
		// value we leave it as false.
		json.Unmarshal(value, &m.exclusive)
	}
	return nil
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

func (m *minimum) SetSchema(v map[string]json.RawMessage) error {
	value, ok := v["exclusiveminimum"]
	if ok {
		// Ignore errors from Unmarshal. If exclusiveminimum is a non boolean JSON
		// value we leave it as false.
		json.Unmarshal(value, &m.exclusive)
	}
	return nil
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

type maxProperties int

func (m maxProperties) Validate(v interface{}) []ValidationError {
	val, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	if len(val) > int(m) {
		return []ValidationError{ValidationError{
			fmt.Sprintf("Object has more properties than maxProperties (%d > %d)", len(val), m)}}
	}
	return nil
}

func (m *maxProperties) UnmarshalJSON(b []byte) error {
	var n int
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	if n < 0 {
		return fmt.Errorf("maxProperties cannot be smaller than zero")
	}
	*m = maxProperties(n)
	return nil
}

type minProperties int

func (m minProperties) Validate(v interface{}) []ValidationError {
	val, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	if len(val) < int(m) {
		return []ValidationError{ValidationError{
			fmt.Sprintf("Object has fewer properties than minProperties (%d < %d)", len(val), m)}}
	}
	return nil
}

func (m *minProperties) UnmarshalJSON(b []byte) error {
	var n int
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	if n < 0 {
		return fmt.Errorf("minProperties cannot be smaller than zero")
	}
	*m = minProperties(n)
	return nil
}

type patternProperties struct {
	object []regexpToSchema
}

type regexpToSchema struct {
	regexp regexp.Regexp
	schema Schema
}

func (p patternProperties) Validate(v interface{}) []ValidationError {
	var valErrs []ValidationError
	data, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	for dataKey, dataVal := range data {
		for _, val := range p.object {
			if val.regexp.MatchString(dataKey) {
				valErrs = append(valErrs, val.schema.Validate(dataVal)...)
			}
		}
	}
	return valErrs
}

func (p *patternProperties) SetSchema(v map[string]json.RawMessage) error {
	if _, ok := v["properties"]; ok {
		return fmt.Errorf("we don't need an independent 'patternProperties' validator, " +
			"because 'properties' is one of this schema key's neighbors and will handle its validation")
	}
	return nil
}

func (p *patternProperties) UnmarshalJSON(b []byte) error {
	var m map[string]Schema
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	for key, val := range m {
		r, err := regexp.Compile(key)
		if err != nil {
			return err
		}
		p.object = append(p.object, regexpToSchema{*r, val})
	}
	return nil
}

type properties struct {
	object                     map[string]Schema
	additionalPropertiesBool   bool
	additionalPropertiesObject *Schema
}

func (p properties) Validate(v interface{}) []ValidationError {
	var valErrs []ValidationError
	dataMap, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	for dataKey, dataValue := range dataMap {
		schema, ok := p.object[dataKey]
		if ok {
			valErrs = append(valErrs, schema.Validate(dataValue)...)
		} else if p.additionalPropertiesObject != nil {
			valErrs = append(valErrs, p.additionalPropertiesObject.Validate(dataValue)...)
		} else if !p.additionalPropertiesBool {
			return []ValidationError{ValidationError{"Additional properties aren't allowed"}}
		}
	}
	return valErrs
}

func (p *properties) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &p.object); err != nil {
		return err
	}
	return nil
}

func (p *properties) SetSchema(v map[string]json.RawMessage) error {
	p.additionalPropertiesBool = true
	val, ok := v["additionalProperties"]
	if !ok {
		return nil
	}
	json.Unmarshal(val, &p.additionalPropertiesBool)
	if err := json.Unmarshal(val, &p.additionalPropertiesObject); err != nil {
		p.additionalPropertiesObject = nil
	}
	return nil
}

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
