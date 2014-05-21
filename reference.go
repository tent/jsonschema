package jsonschema

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Using a pointer allows us to handle recursive embedded schemas.
type EmbeddedSchemas map[string]*Schema

func (e *EmbeddedSchemas) UnmarshalJSON(b []byte) error {
	*e = make(EmbeddedSchemas)
	err1 := e.UnmarshalArray(b)
	err2 := e.UnmarshalObject(b)
	err3 := e.UnmarshalSingle(b)
	if err1 != nil && err2 != nil && err3 != nil {
		return errors.New("no valid embedded schemas")
	}
	return nil
}

func (e *EmbeddedSchemas) UnmarshalArray(b []byte) error {
	var schemas []*Schema
	if err := json.Unmarshal(b, &schemas); err != nil {
		return err
	}
	for i, v := range schemas {
		(*e)[strconv.Itoa(i)] = v
	}
	return nil
}

func (e *EmbeddedSchemas) UnmarshalObject(b []byte) error {
	var m map[string]*Schema
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	for k, v := range m {
		(*e)[k] = v
	}
	return nil
}

func (e *EmbeddedSchemas) UnmarshalSingle(b []byte) error {
	var s Schema
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	(*e)[""] = &s
	return nil
}

// resolveRefs starts a depth-first search through a document for schemas containing
// the 'ref' validator. It completely resolves each one found.
func (s *Schema) resolveRefs() {
	s.resolveSelfAndBelow(*s)
}

func (s *Schema) resolveSelfAndBelow(rootSchema Schema) {
	s.resolveSelf(rootSchema)
	s.resolveBelow(rootSchema)
}

func (s *Schema) resolveSelf(rootSchema Schema) {
	if str, ok := s.hasRef(); ok {
		sch, err := refToSchema(str, rootSchema)
		if err != nil {
			return
		}
		*s = *sch
		s.resolveSelf(rootSchema)
	}
}

// TODO: test that we fail gracefully if the schema contains infinitely looping "$ref"s.
func (s *Schema) resolveBelow(rootSchema Schema) {
	if s.resolved == true {
		return
	}
	s.resolved = true
	for _, n := range s.nodes {
		for _, sch := range n.EmbeddedSchemas {
			sch.resolveSelfAndBelow(rootSchema)
		}
	}
}

func (s *Schema) hasRef() (string, bool) {
	for _, n := range s.nodes {
		if r, ok := n.Validator.(*ref); ok {
			return string(*r), true
		}
	}
	return "", false
}

func refToSchema(str string, rootSchema Schema) (*Schema, error) {
	var split []string
	url, err := url.Parse(str)
	if err == nil && url.IsAbs() {
		// Handle external URIs.
		if !LoadExternalSchemas {
			return new(Schema), errors.New("external schemas are disabled")
		}
		resp, err := http.Get(str)
		if err != nil {
			return new(Schema), errors.New("bad external url")
		}
		defer resp.Body.Close()
		s, err := Parse(resp.Body)
		if err != nil {
			return new(Schema), errors.New("error parsing external doc")
		}
		str = url.Fragment
		rootSchema = *s
	} else {
		// Remove the prefix from internal URIs.
		if strings.HasPrefix(str, "#/") {
			str = str[2:len(str)]
		} else if strings.HasPrefix(str, "#") {
			str = str[1:len(str)]
		}
	}
	split = strings.Split(str, "/")
	// Make replacements.
	for i, v := range split {
		r := strings.NewReplacer("~0", "~", "~1", "/", "%25", "%")
		split[i] = r.Replace(v)
	}
	// Resolve the local part of the URI.
	return resolveLocalPath(split, rootSchema, str)
}

// TODO: add code and tests for references more than one level deep.
func resolveLocalPath(split []string, rootSchema Schema, str string) (*Schema, error) {
	switch len(split) {
	case 1:
		if split[0] == "" {
			return &rootSchema, nil
		}
		v, ok := rootSchema.nodes[split[0]]
		if ok == false {
			break
		}
		if s, ok := v.EmbeddedSchemas[""]; ok {
			return s, nil
		}
	case 2:
		v, ok := rootSchema.nodes[split[0]]
		if ok == false {
			break
		}
		if s, ok := v.EmbeddedSchemas[split[1]]; ok {
			return s, nil
		}
	}
	return new(Schema), fmt.Errorf("failed to resolve %s", str)
}
