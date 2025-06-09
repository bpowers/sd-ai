package schema

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestType_SingleType(t *testing.T) {
	tests := []struct {
		name         string
		schemaType   Type
		expectedJSON string
	}{
		{
			name:         "string type",
			schemaType:   Type{types: []JsonType{String}},
			expectedJSON: `"string"`,
		},
		{
			name:         "number type",
			schemaType:   Type{types: []JsonType{Number}},
			expectedJSON: `"number"`,
		},
		{
			name:         "object type",
			schemaType:   Type{types: []JsonType{Object}},
			expectedJSON: `"object"`,
		},
		{
			name:         "array type",
			schemaType:   Type{types: []JsonType{Array}},
			expectedJSON: `"array"`,
		},
		{
			name:         "null type",
			schemaType:   Type{types: []JsonType{Null}},
			expectedJSON: `"null"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Test marshaling
			jsonBytes, err := json.Marshal(tc.schemaType)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedJSON, string(jsonBytes))

			// Test unmarshaling
			var unmarshaledType Type
			err = json.Unmarshal(jsonBytes, &unmarshaledType)
			require.NoError(t, err)
			assert.False(t, unmarshaledType.IsUnion())
			assert.Equal(t, tc.schemaType.Types(), unmarshaledType.Types())
		})
	}
}

func TestType_UnionType(t *testing.T) {
	tests := []struct {
		name         string
		schemaType   Type
		expectedJSON string
	}{
		{
			name:         "object or null",
			schemaType:   Type{types: []JsonType{Object, Null}},
			expectedJSON: `["object","null"]`,
		},
		{
			name:         "string or number",
			schemaType:   Type{types: []JsonType{String, Number}},
			expectedJSON: `["string","number"]`,
		},
		{
			name:         "string, number, or null",
			schemaType:   Type{types: []JsonType{String, Number, Null}},
			expectedJSON: `["string","number","null"]`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Test marshaling
			jsonBytes, err := json.Marshal(tc.schemaType)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedJSON, string(jsonBytes))

			// Test unmarshaling
			var unmarshaledType Type
			err = json.Unmarshal(jsonBytes, &unmarshaledType)
			require.NoError(t, err)
			assert.True(t, unmarshaledType.IsUnion())
			assert.ElementsMatch(t, tc.schemaType.Types(), unmarshaledType.Types())
		})
	}
}

func TestType_Contains(t *testing.T) {
	tests := []struct {
		name       string
		schemaType Type
		checkType  JsonType
		expected   bool
	}{
		{
			name:       "single type contains self",
			schemaType: Type{types: []JsonType{String}},
			checkType:  String,
			expected:   true,
		},
		{
			name:       "single type does not contain other",
			schemaType: Type{types: []JsonType{String}},
			checkType:  Number,
			expected:   false,
		},
		{
			name:       "union type contains first",
			schemaType: Type{types: []JsonType{String, Number, Null}},
			checkType:  String,
			expected:   true,
		},
		{
			name:       "union type contains middle",
			schemaType: Type{types: []JsonType{String, Number, Null}},
			checkType:  Number,
			expected:   true,
		},
		{
			name:       "union type contains last",
			schemaType: Type{types: []JsonType{String, Number, Null}},
			checkType:  Null,
			expected:   true,
		},
		{
			name:       "union type does not contain other",
			schemaType: Type{types: []JsonType{String, Number}},
			checkType:  Object,
			expected:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.schemaType.Contains(tc.checkType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestType_String(t *testing.T) {
	tests := []struct {
		name       string
		schemaType Type
		expected   string
	}{
		{
			name:       "single string type",
			schemaType: Type{types: []JsonType{String}},
			expected:   "string",
		},
		{
			name:       "single number type",
			schemaType: Type{types: []JsonType{Number}},
			expected:   "number",
		},
		{
			name:       "union types",
			schemaType: Type{types: []JsonType{Object, Null}},
			expected:   "union[object null]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.schemaType.String()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestType_UnmarshalError(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
	}{
		{
			name:     "invalid type",
			jsonData: `123`,
		},
		{
			name:     "invalid array element",
			jsonData: `["string", 123]`,
		},
		{
			name:     "empty object",
			jsonData: `{}`,
		},
		{
			name:     "empty array",
			jsonData: `[]`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var schemaType Type
			err := json.Unmarshal([]byte(tc.jsonData), &schemaType)
			assert.Error(t, err)
		})
	}
}

func TestJSON_WithSchemaType(t *testing.T) {
	tests := []struct {
		name     string
		schema   JSON
		expected string
	}{
		{
			name: "single type schema",
			schema: JSON{
				Type:        Type{types: []JsonType{String}},
				Description: "A string field",
			},
			expected: `{"type":"string","description":"A string field"}`,
		},
		{
			name: "union type schema",
			schema: JSON{
				Type: Type{types: []JsonType{Object, Null}},
				Description: "An object or null field",
			},
			expected: `{"type":["object","null"],"description":"An object or null field"}`,
		},
		{
			name: "complex schema with union type",
			schema: JSON{
				Type: Type{types: []JsonType{String, Number}},
				Description: "A string or number field",
				Enum:        []string{"foo", "bar"},
			},
			expected: `{"type":["string","number"],"description":"A string or number field","enum":["foo","bar"]}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Test marshaling
			jsonBytes, err := json.Marshal(tc.schema)
			require.NoError(t, err)
			assert.JSONEq(t, tc.expected, string(jsonBytes))

			// Test unmarshaling
			var unmarshaledSchema JSON
			err = json.Unmarshal(jsonBytes, &unmarshaledSchema)
			require.NoError(t, err)

			// Verify the type was correctly unmarshaled
			assert.Equal(t, tc.schema.Type.IsUnion(), unmarshaledSchema.Type.IsUnion())
			assert.ElementsMatch(t, tc.schema.Type.Types(), unmarshaledSchema.Type.Types())
			assert.Equal(t, tc.schema.Description, unmarshaledSchema.Description)
		})
	}
}

func TestType_RegressionRealWorldUsage(t *testing.T) {
	// Test based on the real schema that has union types like ["object", "null"]
	jsonSchema := `{
		"type": "object",
		"properties": {
			"graphicalFunction": {
				"type": ["object", "null"],
				"description": "Optional graphical function"
			},
			"name": {
				"type": "string",
				"description": "Variable name"
			}
		}
	}`

	var schema JSON
	err := json.Unmarshal([]byte(jsonSchema), &schema)
	require.NoError(t, err)

	// Verify the root type
	assert.False(t, schema.Type.IsUnion())
	types := schema.Type.Types()
	assert.Equal(t, []JsonType{Object}, types)

	// Verify the graphicalFunction property has union type
	require.NotNil(t, schema.Properties)
	graphicalFunc, exists := schema.Properties["graphicalFunction"]
	require.True(t, exists)

	assert.True(t, graphicalFunc.Type.IsUnion())
	assert.ElementsMatch(t, []JsonType{Object, Null}, graphicalFunc.Type.Types())
	assert.True(t, graphicalFunc.Type.Contains(Object))
	assert.True(t, graphicalFunc.Type.Contains(Null))
	assert.False(t, graphicalFunc.Type.Contains(String))

	// Verify the name property has single type
	name, exists := schema.Properties["name"]
	require.True(t, exists)

	assert.False(t, name.Type.IsUnion())
	nameTypes := name.Type.Types()
	assert.Equal(t, []JsonType{String}, nameTypes)

	// Test round-trip marshaling
	marshaled, err := json.Marshal(schema)
	require.NoError(t, err)

	var roundTrip JSON
	err = json.Unmarshal(marshaled, &roundTrip)
	require.NoError(t, err)

	// Verify round-trip preserves union types
	graphicalFuncRT := roundTrip.Properties["graphicalFunction"]
	assert.True(t, graphicalFuncRT.Type.IsUnion())
	assert.ElementsMatch(t, []JsonType{Object, Null}, graphicalFuncRT.Type.Types())
}

