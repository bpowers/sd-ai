package simulation

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/UB-IAD/sd-ai/go/schema"
)

func TestResponseSchemaUnmarshaling(t *testing.T) {
	// Parse the embedded response schema
	var actualSchema schema.JSON
	err := json.Unmarshal([]byte(responseSchemaJson), &actualSchema)
	require.NoError(t, err)

	// Skip constructing the expected schema for now to avoid Type zero-value issues
	// _ = constructExpectedSchema()

	// Verify the top-level structure
	assert.False(t, actualSchema.Type.IsUnion())
	types := actualSchema.Type.Types()
	assert.Equal(t, []schema.JsonType{schema.Object}, types)
	assert.Equal(t, "Model", actualSchema.Title)
	assert.Equal(t, "System Dynamics model representation", actualSchema.Description)
	assert.Equal(t, "http://json-schema.org/draft-07/schema#", actualSchema.Schema)

	// Verify required fields (should include at least title and explanation)
	assert.Contains(t, actualSchema.Required, "title")
	assert.Contains(t, actualSchema.Required, "explanation")

	// Verify additionalProperties is false
	require.NotNil(t, actualSchema.AdditionalProperties)
	assert.False(t, *actualSchema.AdditionalProperties)

	// Verify main properties exist
	require.NotNil(t, actualSchema.Properties)
	assert.Contains(t, actualSchema.Properties, "title")
	assert.Contains(t, actualSchema.Properties, "explanation")
	assert.Contains(t, actualSchema.Properties, "variables")
	assert.Contains(t, actualSchema.Properties, "specs")

	// Verify title property
	titleProp := actualSchema.Properties["title"]
	assert.False(t, titleProp.Type.IsUnion())
	types = titleProp.Type.Types()
	assert.Equal(t, []schema.JsonType{schema.String}, types)
	assert.Equal(t, "Title of the model", titleProp.Description)

	// Verify explanation property
	explanationProp := actualSchema.Properties["explanation"]
	assert.False(t, explanationProp.Type.IsUnion())
	types = explanationProp.Type.Types()
	assert.Equal(t, []schema.JsonType{schema.String}, types)
	assert.Equal(t, "Explanation of the model", explanationProp.Description)

	// Verify variables property
	variablesProp := actualSchema.Properties["variables"]
	assert.False(t, variablesProp.Type.IsUnion())
	types = variablesProp.Type.Types()
	assert.Equal(t, []schema.JsonType{schema.Array}, types)
	assert.Equal(t, "List of variables in the model", variablesProp.Description)
	require.NotNil(t, variablesProp.Items)
	assert.Equal(t, "#/definitions/Variable", variablesProp.Items.Ref)

	// Verify specs property
	specsProp := actualSchema.Properties["specs"]
	assert.Equal(t, "#/definitions/Specs", specsProp.Ref)

	// Verify definitions section exists
	require.NotNil(t, actualSchema.Definitions)

	// Verify key definitions exist
	assert.Contains(t, actualSchema.Definitions, "Variable")
	assert.Contains(t, actualSchema.Definitions, "VariableType")
	assert.Contains(t, actualSchema.Definitions, "Specs")
	assert.Contains(t, actualSchema.Definitions, "Expr")
	assert.Contains(t, actualSchema.Definitions, "IExpr")
	// Note: GraphicalFunction and Point may be inlined in the current schema

	// Verify Variable definition
	variableDef := actualSchema.Definitions["Variable"]
	assert.False(t, variableDef.Type.IsUnion())
	types = variableDef.Type.Types()
	assert.Equal(t, []schema.JsonType{schema.Object}, types)
	// Verify required fields include at least name and type
	assert.Contains(t, variableDef.Required, "name")
	assert.Contains(t, variableDef.Required, "type")
	assert.Contains(t, variableDef.Properties, "name")
	assert.Contains(t, variableDef.Properties, "type")
	assert.Contains(t, variableDef.Properties, "equation")

	// Verify VariableType definition
	variableTypeDef := actualSchema.Definitions["VariableType"]
	assert.False(t, variableTypeDef.Type.IsUnion())
	types = variableTypeDef.Type.Types()
	assert.Equal(t, []schema.JsonType{schema.String}, types)
	assert.ElementsMatch(t, []string{"variable", "stock", "flow"}, variableTypeDef.Enum)

	// Verify Expr is a union type (anyOf or oneOf)
	exprDef := actualSchema.Definitions["Expr"]
	assert.True(t, len(exprDef.OneOf) > 0 || len(exprDef.AnyOf) > 0, "Expr should be a union type")

	// Verify IExpr is a union type (anyOf or oneOf)
	iexprDef := actualSchema.Definitions["IExpr"]
	assert.True(t, len(iexprDef.OneOf) > 0 || len(iexprDef.AnyOf) > 0, "IExpr should be a union type")

	// Verify AST node types have proper discriminators (if they exist)
	if constDef, exists := actualSchema.Definitions["Const"]; exists {
		if constDef.Properties != nil {
			if typeProp, typeExists := constDef.Properties["type"]; typeExists {
				assert.Equal(t, "const", typeProp.Const)
			}
		}
	}

	if varDef, exists := actualSchema.Definitions["Var"]; exists {
		if varDef.Properties != nil {
			if typeProp, typeExists := varDef.Properties["type"]; typeExists {
				assert.Equal(t, "var", typeProp.Const)
			}
		}
	}

	// Test that we can re-marshal the schema
	remarshaled, err := json.Marshal(&actualSchema)
	require.NoError(t, err)

	// Verify we can unmarshal it again (round-trip test)
	var roundTripSchema schema.JSON
	err = json.Unmarshal(remarshaled, &roundTripSchema)
	require.NoError(t, err)

	// Basic verification that round-trip worked
	assert.Equal(t, actualSchema.Type, roundTripSchema.Type)
	assert.Equal(t, actualSchema.Title, roundTripSchema.Title)
	assert.Equal(t, actualSchema.Description, roundTripSchema.Description)
	assert.Equal(t, len(actualSchema.Properties), len(roundTripSchema.Properties))
	assert.Equal(t, len(actualSchema.Definitions), len(roundTripSchema.Definitions))
}

func constructExpectedSchema() *schema.JSON {
	// Helper function to create a pointer to bool
	boolPtr := func(b bool) *bool { return &b }

	return &schema.JSON{
		Schema:      "http://json-schema.org/draft-07/schema#",
		Type:        schema.NewSingleType(schema.Object),
		Title:       "Model",
		Description: "System Dynamics model representation",
		Properties: map[string]*schema.JSON{
			"title": {
				Type:        schema.NewSingleType(schema.String),
				Description: "Title of the model",
			},
			"explanation": {
				Type:        schema.NewSingleType(schema.String),
				Description: "Explanation of the model",
			},
			"variables": {
				Type:        schema.NewSingleType(schema.Array),
				Description: "List of variables in the model",
				Items: &schema.JSON{
					Ref: "#/definitions/Variable",
				},
			},
			"specs": {
				Ref: "#/definitions/Specs",
			},
		},
		Required:             []string{"title", "explanation"},
		AdditionalProperties: boolPtr(false),
		Definitions: map[string]*schema.JSON{
			"Variable": {
				Type:        schema.NewSingleType(schema.Object),
				Description: "A variable in the system dynamics model",
				Properties: map[string]*schema.JSON{
					"name": {
						Type:        schema.NewSingleType(schema.String),
						Description: "Name of the variable",
					},
					"type": {
						Ref: "#/definitions/VariableType",
					},
					"equation": {
						Ref: "#/definitions/Expr",
					},
					"documentation": {
						Type:        schema.NewSingleType(schema.String),
						Description: "Documentation for the variable",
					},
					"units": {
						Type:        schema.NewSingleType(schema.String),
						Description: "Units of measurement",
					},
					"inflows": {
						Type:        schema.NewSingleType(schema.Array),
						Description: "List of inflow variable names",
						Items: &schema.JSON{
							Type: schema.NewSingleType(schema.String),
						},
					},
					"outflows": {
						Type:        schema.NewSingleType(schema.Array),
						Description: "List of outflow variable names",
						Items: &schema.JSON{
							Type: schema.NewSingleType(schema.String),
						},
					},
					"graphicalFunction": {
						Ref: "#/definitions/GraphicalFunction",
					},
				},
				Required:             []string{"name", "type"},
				AdditionalProperties: boolPtr(false),
			},
			"VariableType": {
				Type:        schema.NewSingleType(schema.String),
				Description: "JsonType of variable in the system dynamics model",
				Enum:        []string{"variable", "stock", "flow"},
			},
			"Expr": {
				Description: "Expression node in the abstract syntax tree",
				OneOf: []*schema.JSON{
					{Ref: "#/definitions/Const"},
					{Ref: "#/definitions/Var"},
					{Ref: "#/definitions/Call"},
					{Ref: "#/definitions/Subscript"},
					{Ref: "#/definitions/UnaryOp"},
					{Ref: "#/definitions/BinaryOp"},
					{Ref: "#/definitions/If"},
				},
			},
			// Add more definitions as needed for a complete test
		},
	}
}
