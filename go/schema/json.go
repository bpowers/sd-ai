package schema

import (
	"encoding/json"
	"fmt"
)

const URL = "http://json-schema.org/draft-07/schema#"

type JsonType string

const (
	String JsonType = "string"
	Array  JsonType = "array"
	Object JsonType = "object"
	Number JsonType = "number"
	Null   JsonType = "null"
)

// Type represents a JSON Schema type that can be either a single type
// or a union of types
type Type struct {
	types []JsonType
}

// NewSingleType creates a Type with a single type
func NewSingleType(t JsonType) Type {
	return Type{types: []JsonType{t}}
}

// NewUnionType creates a Type with multiple types (union)
func NewUnionType(types ...JsonType) Type {
	return Type{types: types}
}

// IsUnion returns true if this is a union type (has more than one type)
func (t Type) IsUnion() bool {
	return len(t.types) > 1
}

// Types returns a copy of the underlying types slice
func (t Type) Types() []JsonType {
	result := make([]JsonType, len(t.types))
	copy(result, t.types)
	return result
}

// IsZero returns true if the Type has no types (for omitzero behavior)
func (t Type) IsZero() bool {
	return len(t.types) == 0
}

// MarshalJSON implements json.Marshaler for Type
func (t Type) MarshalJSON() ([]byte, error) {
	if len(t.types) == 0 {
		// For empty types (used in oneOf/anyOf cases), return null which will be omitted
		return []byte("null"), nil
	}
	if len(t.types) == 1 {
		return json.Marshal(t.types[0])
	}
	return json.Marshal(t.types)
}

// UnmarshalJSON implements json.Unmarshaler for Type
func (t *Type) UnmarshalJSON(data []byte) error {
	// Handle null values (for oneOf/anyOf cases)
	if string(data) == "null" {
		t.types = nil
		return nil
	}

	// Try to unmarshal as array first
	var types []JsonType
	if err := json.Unmarshal(data, &types); err == nil {
		if len(types) == 0 {
			return fmt.Errorf("type array cannot be empty")
		}
		t.types = types
		return nil
	}

	// Try to unmarshal as single type
	var singleType JsonType
	if err := json.Unmarshal(data, &singleType); err == nil {
		t.types = []JsonType{singleType}
		return nil
	}

	return fmt.Errorf("cannot unmarshal %s as Type", string(data))
}

// String returns a string representation of the Type
func (t Type) String() string {
	if len(t.types) == 0 {
		return "invalid"
	}
	if len(t.types) == 1 {
		return string(t.types[0])
	}
	return fmt.Sprintf("union%v", t.types)
}

// Contains checks if the Type contains a specific type
func (t Type) Contains(jt JsonType) bool {
	for _, typ := range t.types {
		if typ == jt {
			return true
		}
	}
	return false
}

// JSON is a way to describe a JSON Schema
type JSON struct {
	Type                 Type             `json:"type,omitzero"`
	Title                string           `json:"title,omitzero"`
	Description          string           `json:"description,omitzero"`
	Properties           map[string]*JSON `json:"properties,omitzero"`
	Items                *JSON            `json:"items,omitzero"`
	Enum                 []string         `json:"enum,omitzero"`
	Required             []string         `json:"required,omitzero"`
	AdditionalProperties *bool            `json:"additionalProperties,omitzero"`
	Schema               string           `json:"$schema,omitzero"`
	OneOf                []*JSON          `json:"oneOf,omitzero"`
	AnyOf                []*JSON          `json:"anyOf,omitzero"`
	Ref                  string           `json:"$ref,omitzero"`
	Const                string           `json:"const,omitzero"`
	Definitions          map[string]*JSON `json:"definitions,omitzero"`
}
