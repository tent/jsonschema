package jsonschema

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
)

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
	if s.resolved == true {
		return
	}
	s.resolved = true
	for _, n := range s.nodes {
		for _, sch := range n.EmbeddedSchemas {
			sch.resolveItselfAndBelow(rootSchema)
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
		v, ok := rootSchema.nodes[split[0]]
		if ok == false {
			return new(Schema), errors.New("resolve failed (len 2)")
		}
		s2, ok := v.EmbeddedSchemas[""]
		if ok {
			return s2, nil
		}
	case len(split) == 2:
		v, ok := rootSchema.nodes[split[0]]
		if ok == false {
			return new(Schema), errors.New("resolve failed (len 2)")
		}
		s2, ok := v.EmbeddedSchemas[split[1]]
		if ok {
			return s2, nil
		}
	}
	return new(Schema), errors.New("resolve failed (other)")
}
