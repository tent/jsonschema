package jsonschema

// A temporary interface for use until code is written to access
// a validator's EmbeddedSchemas field with reflection.
type SchemaEmbedder interface {
	LinkEmbedded(map[string]*Schema)
}

func (a *other) LinkEmbedded(b map[string]*Schema) {
	a.EmbeddedSchemas = b
}

func (a *additionalItems) LinkEmbedded(b map[string]*Schema) {
	a.EmbeddedSchemas = b
}

func (a *items) LinkEmbedded(b map[string]*Schema) {
	a.EmbeddedSchemas = b
}

func (a *dependencies) LinkEmbedded(b map[string]*Schema) {
	a.EmbeddedSchemas = b
}

func (a *additionalProperties) LinkEmbedded(b map[string]*Schema) {
	a.EmbeddedSchemas = b
}

func (a *patternProperties) LinkEmbedded(b map[string]*Schema) {
	a.EmbeddedSchemas = b
}

func (a *properties) LinkEmbedded(b map[string]*Schema) {
	a.EmbeddedSchemas = b
}

func (a *allOf) LinkEmbedded(b map[string]*Schema) {
	a.EmbeddedSchemas = b
}

func (a *anyOf) LinkEmbedded(b map[string]*Schema) {
	a.EmbeddedSchemas = b
}

func (a *definitions) LinkEmbedded(b map[string]*Schema) {
	a.EmbeddedSchemas = b
}

func (a *not) LinkEmbedded(b map[string]*Schema) {
	a.EmbeddedSchemas = b
}

func (a *oneOf) LinkEmbedded(b map[string]*Schema) {
	a.EmbeddedSchemas = b
}
