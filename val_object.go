package jsonschema

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

type additionalProperties struct {
	isTrue               bool
	sch                  *Schema
	propertiesIsNeighbor bool
}

func (a *additionalProperties) UnmarshalJSON(b []byte) error {
	a.isTrue = true
	if err := json.Unmarshal(b, &a.isTrue); err == nil {
		return nil
	}
	if err := json.Unmarshal(b, &a.sch); err != nil {
		a.sch = nil
		return err
	}
	return nil
}

func (a *additionalProperties) CheckNeighbors(m map[string]Node) {
	v, ok := m["properties"]
	if !ok {
		return
	}
	if _, ok := v.Validator.(*properties); !ok {
		return
	}
	a.propertiesIsNeighbor = true
	return
}

func (a additionalProperties) Validate(v interface{}) []ValidationError {
	// In this case validation will be handled by the "properties" validator.
	if a.propertiesIsNeighbor {
		return nil
	}
	var valErrs []ValidationError
	dataMap, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	for _, dataVal := range dataMap {
		valErrs = append(valErrs, a.sch.Validate(dataVal)...)
	}
	return valErrs
}

type dependencies struct {
	schemaDeps   map[string]Schema
	propertyDeps map[string]propertySet
}

type propertySet map[string]struct{}

func (d *dependencies) UnmarshalJSON(b []byte) error {
	var c map[string]json.RawMessage
	if err := json.Unmarshal(b, &c); err != nil {
		return err
	}

	d.schemaDeps = make(map[string]Schema, len(c))
	for k, v := range c {
		var s Schema
		if err := json.Unmarshal(v, &s); err != nil {
			continue
		}
		d.schemaDeps[k] = s
	}

	d.propertyDeps = make(map[string]propertySet, len(c))
	for k, v := range c {
		var props []string
		if err := json.Unmarshal(v, &props); err != nil {
			continue
		}
		set := make(propertySet, len(props))
		for _, p := range props {
			set[p] = struct{}{}
		}
		d.propertyDeps[k] = set
	}

	if len(d.propertyDeps) == 0 && len(d.schemaDeps) == 0 {
		return errors.New("no valid schema or property dependency validators")
	}
	return nil
}

func (d dependencies) Validate(v interface{}) []ValidationError {
	var valErrs []ValidationError
	val, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}

	// Handle schema dependencies.
	for key, schema := range d.schemaDeps {
		if _, ok := val[key]; !ok {
			continue
		}
		valErrs = append(valErrs, schema.Validate(v)...)
	}

	// Handle property dependencies.
	for key, set := range d.propertyDeps {
		if _, ok := val[key]; !ok {
			continue
		}
		for a := range set {
			if _, ok := val[a]; !ok {
				valErrs = append(valErrs, ValidationError{
					fmt.Sprintf("instance does not have a property with the name %s", a)})
			}
		}
	}

	return valErrs
}

type maxProperties int

func (m *maxProperties) UnmarshalJSON(b []byte) error {
	var n int
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	if n < 0 {
		return errors.New("maxProperties cannot be smaller than zero")
	}
	*m = maxProperties(n)
	return nil
}

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

type minProperties int

func (m *minProperties) UnmarshalJSON(b []byte) error {
	var n int
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	if n < 0 {
		return errors.New("minProperties cannot be smaller than zero")
	}
	*m = minProperties(n)
	return nil
}

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

type patternProperties struct {
	object   []regexpToSchema
	disabled bool
}

type regexpToSchema struct {
	regexp regexp.Regexp
	schema Schema
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

func (p *patternProperties) CheckNeighbors(m map[string]Node) {
	v, ok := m["properties"]
	if !ok {
		return
	}
	if _, ok := v.Validator.(*properties); !ok {
		return
	}
	p.disabled = true
	return
}

func (p patternProperties) Validate(v interface{}) []ValidationError {
	if p.disabled {
		return nil
	}
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

type properties struct {
	object                     map[string]Schema
	patternProperties          *patternProperties
	additionalPropertiesBool   bool
	additionalPropertiesObject *Schema
}

func (p *properties) UnmarshalJSON(b []byte) error {
	p.additionalPropertiesBool = true
	return json.Unmarshal(b, &p.object)
}

func (p *properties) CheckNeighbors(m map[string]Node) {
	v, ok := m["patternProperties"]
	if ok {
		pat, ok := v.Validator.(*patternProperties)
		if ok {
			p.patternProperties = &patternProperties{pat.object, false}
		}
	}
	v, ok = m["additionalProperties"]
	if !ok {
		return
	}
	a, ok := v.Validator.(*additionalProperties)
	if !ok {
		return
	}
	p.additionalPropertiesBool = a.isTrue
	p.additionalPropertiesObject = a.sch
	return
}

func (p properties) Validate(v interface{}) []ValidationError {
	var valErrs []ValidationError
	dataMap, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	for dataKey, dataVal := range dataMap {
		var match = false
		schema, ok := p.object[dataKey]
		if ok {
			valErrs = append(valErrs, schema.Validate(dataVal)...)
			match = true
		}
		if p.patternProperties != nil {
			for _, val := range p.patternProperties.object {
				if val.regexp.MatchString(dataKey) {
					valErrs = append(valErrs, val.schema.Validate(dataVal)...)
					match = true
				}
			}
		}
		if match {
			continue
		}
		if p.additionalPropertiesObject != nil {
			valErrs = append(valErrs, p.additionalPropertiesObject.Validate(dataVal)...)
			continue
		}
		if !p.additionalPropertiesBool {
			valErrs = append([]ValidationError{ValidationError{"Additional properties aren't allowed"}})
		}
	}
	return valErrs
}

type required map[string]struct{}

func (r *required) UnmarshalJSON(b []byte) error {
	var l []string
	if err := json.Unmarshal(b, &l); err != nil {
		return err
	}
	*r = make(required)
	for _, val := range l {
		(*r)[val] = struct{}{}
	}
	return nil
}

func (r required) Validate(v interface{}) []ValidationError {
	var valErrs []ValidationError
	data, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	for key := range r {
		if _, ok := data[key]; !ok {
			valErrs = append(valErrs, ValidationError{fmt.Sprintf("Required error. The data must be an object with %v as one of its keys", key)})
		}
	}
	return valErrs
}
