package simulation

import (
	_ "embed"
	"encoding/json"

	"github.com/UB-IAD/sd-ai/go/ast"
	"github.com/UB-IAD/sd-ai/go/schema"
	"github.com/UB-IAD/sd-ai/go/sdjson"
)

//go:embed response_schema.json
var responseSchemaJson string

var RelationshipsResponseSchema *schema.JSON

func init() {
	RelationshipsResponseSchema = new(schema.JSON)
	err := json.Unmarshal([]byte(responseSchemaJson), RelationshipsResponseSchema)
	if err != nil {
		panic(err)
	}
}

type Variable struct {
	Name              string                    `json:"name"`
	Type              sdjson.VariableType       `json:"type"`
	Equation          *ast.ExprJSON             `json:"equation,omitempty"`
	Documentation     string                    `json:"documentation,omitempty"`
	Units             string                    `json:"units,omitempty"`
	Inflows           []string                  `json:"inflows,omitempty"`
	Outflows          []string                  `json:"outflows,omitempty"`
	GraphicalFunction *sdjson.GraphicalFunction `json:"graphicalFunction,omitempty"`
}

type RelationshipEntry struct {
	Variable          string `json:"variable"`
	Polarity          string `json:"polarity"` // "+", or "-"
	PolarityReasoning string `json:"polarity_reasoning"`
}

type Chain struct {
	InitialVariable string              `json:"initial_variable"`
	Relationships   []RelationshipEntry `json:"relationships"`
	Reasoning       string              `json:"reasoning"`
}

type Model struct {
	Title       string       `json:"title"`
	Explanation string       `json:"explanation"`
	Variables   []Variable   `json:"variables,omitempty"`
	Specs       sdjson.Specs `json:"specs,omitempty"`
}

func (m *Model) Compat() sdjson.Model {
	mdl := sdjson.Model{
		Specs:     m.Specs,
		Variables: make([]sdjson.Variable, 0, len(m.Variables)),
	}

	for _, v := range m.Variables {
		mdl.Variables = append(mdl.Variables,
			sdjson.Variable{
				Name:              v.Name,
				Type:              v.Type,
				Equation:          string(v.Equation.AppendEquation(nil)),
				Documentation:     v.Documentation,
				Units:             v.Units,
				Inflows:           v.Inflows,
				Outflows:          v.Outflows,
				GraphicalFunction: v.GraphicalFunction,
			},
		)
	}

	return mdl
}
