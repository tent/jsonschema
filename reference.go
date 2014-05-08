package jsonschema

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// A SchemaContainer is a schema that both (a) can contain embedded schemas
// and (b) supports resolving "$ref"s to and from those embedded schemas.
type SchemaContainer interface {
	EmbeddedSchemas() []*Schema
	GetSchema(string) *Schema
}

// resolveRefs starts a depth-first search through a document for schemas containing
// the 'ref' validator. It completely resolves each one found.
func (s *Schema) resolveRefs() {
	s.resolveItselfAndBelow(*s)
}

func (s *Schema) resolveItselfAndBelow(rootSchema Schema) {
	s.resolveItself(rootSchema)
	s.resolveBelow(rootSchema)
}

func (s *Schema) resolveItself(rootSchema Schema) {
	if str, ok := s.hasRef(); ok {
		sch, err := refToSchema(str, rootSchema)
		if err != nil {
			return
		}
		*s = *sch
		s.resolveItself(rootSchema)
	}
}

// TODO: test that we fail gracefully if the schema contains infinitely looping "$ref"s.
func (s *Schema) resolveBelow(rootSchema Schema) {
	log.Printf("resolveBelow: %s", s)
	if s.resolved == true {
		return
	}
	s.resolved = true
	for _, v := range s.vals {
		if validator, ok := v.(SchemaContainer); ok {
			for _, a := range validator.EmbeddedSchemas() {
				a.resolveItselfAndBelow(rootSchema)
			}
		}
	}
}

func (s *Schema) hasRef() (string, bool) {
	for _, v := range s.vals {
		if r, ok := v.(*ref); ok {
			return string(*r), true
		}
	}
	return "", false
}

// TODO: improve code, error messages.
func refToSchema(str string, rootSchema Schema) (*Schema, error) {
	var split []string
	url, err := url.Parse(str)
	if err == nil && url.IsAbs() {
		resp, err := http.Get(str)
		if err != nil {
			return new(Schema), errors.New("bad external url")
		}
		defer resp.Body.Close()
		rS, err := Parse(resp.Body)
		if err != nil {
			return new(Schema), errors.New("error parsing external doc")
		}

		str = url.Fragment
		rootSchema = *rS

	} else {
		if strings.HasPrefix(str, "#/") {
			str = str[2:len(str)]
		} else if strings.HasPrefix(str, "#") {
			str = str[1:len(str)]
		}
	}

	split = strings.Split(str, "/")

	for i, v := range split {
		r := strings.NewReplacer("~0", "~", "~1", "/", "%25", "%")
		split[i] = r.Replace(v)
	}

	switch {
	case str == "":
		return &rootSchema, nil
	case len(split) == 1:
		v, ok := rootSchema.vals[split[0]].(SchemaContainer)
		if ok == false {
			return new(Schema), errors.New("resolve failed (len 1)")
		}
		s2 := v.GetSchema("")
		if s2 != nil {
			return s2, nil
		}
	case len(split) == 2:
		v, ok := rootSchema.vals[split[0]].(SchemaContainer)
		if ok == false {
			return new(Schema), errors.New("resolve failed (len 2)")
		}
		s2 := v.GetSchema(split[1])
		if s2 != nil {
			return s2, nil
		}
	}
	return new(Schema), errors.New("resolve failed (other)")
}
