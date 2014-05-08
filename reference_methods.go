package jsonschema

import "strconv"

// See `reference.go` for an explanation of these methods.

// Assume all GetSchema methods are untested. Untested EmbeddedSchemas are marked with a TODO.

func (o other) EmbeddedSchemas() []*Schema {
	s := Schema(o)
	return []*Schema{&s}
}

func (o other) GetSchema(str string) *Schema {
	if str != "" {
		return nil
	}
	s := Schema(o)
	return &s
}

func (i items) EmbeddedSchemas() []*Schema {
	if i.schema != nil {
		return []*Schema{i.schema}
	}
	return i.schemaSlice
}

func (i items) GetSchema(str string) *Schema {
	n, err := strconv.Atoi(str)
	if err != nil {
		return nil
	}
	if n > len(i.schemaSlice)-1 {
		return nil
	}
	return i.schemaSlice[n]
}

// TODO: untested.
func (d dependencies) EmbeddedSchemas() []*Schema {
	schemas := make([]*Schema, len(d.schemaDeps))
	var i int
	for _, v := range d.schemaDeps {
		schemas[i] = &v
		i++
	}
	return schemas
}

func (d dependencies) GetSchema(str string) *Schema {
	sch, ok := d.schemaDeps[str]
	if !ok {
		return nil
	}
	return &sch
}

func (p patternProperties) EmbeddedSchemas() []*Schema {
	schemas := make([]*Schema, len(p.object))
	var i int
	for _, v := range p.object {
		schemas[i] = &v.schema
		i++
	}
	return schemas
}

func (p patternProperties) GetSchema(str string) *Schema {
	v, ok := p.object[str]
	if !ok {
		return nil
	}
	return &v.schema
}

func (p properties) EmbeddedSchemas() []*Schema {
	schemas := make([]*Schema, len(p.object))
	var i int
	for _, v := range p.object {
		schemas[i] = v
		i++
	}
	return schemas
}

func (p properties) GetSchema(str string) *Schema {
	sch, ok := p.object[str]
	if !ok {
		return nil
	}
	return sch
}

func (a allOf) EmbeddedSchemas() []*Schema {
	schemas := make([]*Schema, len(a))
	for i, _ := range a {
		schemas[i] = &a[i]
	}
	return schemas
}

func (a allOf) GetSchema(str string) *Schema {
	n, err := strconv.Atoi(str)
	if err != nil {
		return nil
	}
	if n > len(a)-1 {
		return nil
	}
	return &a[n]
}

// TODO: untested.
func (a anyOf) EmbeddedSchemas() []*Schema {
	schemas := make([]*Schema, len(a))
	for i, _ := range a {
		schemas[i] = &a[i]
	}
	return schemas
}

func (a anyOf) GetSchema(str string) *Schema {
	n, err := strconv.Atoi(str)
	if err != nil {
		return nil
	}
	if n > len(a)-1 {
		return nil
	}
	return &a[n]
}

func (d definitions) EmbeddedSchemas() []*Schema {
	schemas := make([]*Schema, len(d))
	var i int
	for _, v := range d {
		schemas[i] = v
		i++
	}
	return schemas
}

func (d definitions) GetSchema(str string) *Schema {
	sch, ok := d[str]
	if !ok {
		return nil
	}
	return sch
}

// TODO: untested.
func (n not) EmbeddedSchemas() []*Schema {
	return []*Schema{&n.Schema}
}

func (n not) GetSchema(str string) *Schema {
	if str != "" {
		return nil
	}
	return &n.Schema
}

// TODO: untested.
func (o oneOf) EmbeddedSchemas() []*Schema {
	schemas := make([]*Schema, len(o))
	for i, _ := range o {
		schemas[i] = &o[i]
	}
	return schemas
}

func (o oneOf) GetSchema(str string) *Schema {
	n, err := strconv.Atoi(str)
	if err != nil {
		return nil
	}
	if n > len(o)-1 {
		return nil
	}
	return &o[n]
}
