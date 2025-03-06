package schema

const URL = "http://json-schema.org/draft-07/schema#"

type Type string

const (
	String Type = "string"
	Array  Type = "array"
	Object Type = "object"
)

// JSON is a way to describe a JSON Schema
type JSON struct {
	Type                 Type             `json:"type"`
	Description          string           `json:"description,omitempty"`
	Properties           map[string]*JSON `json:"properties,omitempty"`
	Items                *JSON            `json:"items,omitempty"`
	Enum                 []string         `json:"enum,omitempty"`
	Required             []string         `json:"required,omitempty"`
	AdditionalProperties bool             `json:"additionalProperties,omitempty"`
	Schema               string           `json:"$schema,omitempty"`
}
