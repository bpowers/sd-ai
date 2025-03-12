package causal

import (
	"cmp"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/isee-systems/sd-ai/schema"
)

type Set[T cmp.Ordered] map[T]struct{}

func (s Set[T]) Add(e T) {
	s[e] = struct{}{}
}

func (s Set[T]) Contains(e T) bool {
	_, ok := s[e]
	return ok
}

func (s Set[T]) Slice() []T {
	slice := make([]T, 0, len(s))
	for e := range s {
		slice = append(slice, e)
	}
	slices.Sort(slice)

	return slice
}

func NewSet[T cmp.Ordered](elements ...T) Set[T] {
	s := make(Set[T], len(elements))
	for _, element := range elements {
		s.Add(element)
	}
	return s
}

var RelationshipsResponseSchema = &schema.JSON{
	Type: schema.Object,
	Properties: map[string]*schema.JSON{
		"title": {
			Type:        schema.String,
			Description: "A highly descriptive title describing your explanation, with a maximum of 7 words.",
		},
		"explanation": {
			Type:        schema.String,
			Description: "Concisely explain your reasoning for each change you made to the old CLD to create the new CLD. Speak in plain English, don't reference JSON specifically. Don't reiterate the request or any of these instructions.",
		},
		"causal_chains": {
			Type: schema.Array,
			Items: &schema.JSON{
				Type: schema.Object,
				Properties: map[string]*schema.JSON{
					"initial_variable": {
						Type:        schema.String,
						Description: "The first variable in this causal chain.",
					},
					"relationships": {
						Type: schema.Array,
						Items: &schema.JSON{
							Type: schema.Object,
							Properties: map[string]*schema.JSON{
								"variable": {
									Type:        schema.String,
									Description: "A variable in this causal chain.  It is directly influenced by the previous variable in the parent array, and directly influences the next variable in the parent array (if one exists).",
								},
								"polarity": {
									Type: schema.String,
									Enum: []string{
										"+",
										"-",
									},
									Description: "Polarity is either + (positive) or - (negative).  In relationships with positive polarity (+), a change in the previous variable causes a change in the same direction in the current variable.  In relationships with negative polarity (-), an increase in the previous variable causes a decrease in the current variable, and a decrease in the previous variable would cause the current variable to increase.",
								},
								"polarity_reasoning": {
									Type:        schema.String,
									Description: "This is the reason for why the polarity for this relationship was choosen",
								},
							},
							Required: []string{
								"variable",
								"polarity",
								"polarity_reasoning",
							},
							AdditionalProperties: false,
							Description:          "This named variable is influenced by the previous variable, with a given polarity.",
						},
						Description: "Each entry identifies a causal relationship between the previous variable and the current variable, or in the case of the first entry in the array a causal relationship between the variable named in the initial_variable field and the first variable.  If this causal chain represents a feedback loop, the final variable in the chain MUST be the same (have the same name) as the initial_variable.  Every entry in a causal chain MUST have a distinct variable name.",
					},
					"reasoning": {
						Type:        schema.String,
						Description: "This is an explanation for why this causal chain exists.  If it represents a feedback loop, use the words \"feedback loop\", and if it does not represent a feedback loop don't use that term.",
					},
				},
				Required: []string{
					"initial_variable",
					"relationships",
					"reasoning",
				},
				AdditionalProperties: false,
				Description:          "This is a relationship between two variables, from and to (from is the cause, to is the effect).  The relationship also contains a polarity which describes how a change in the from variable impacts the to variable",
			},
			Description: "The list of relationships you think are appropriate to satisfy my request based on all of the information I have given you",
		},
	},
	Required: []string{
		"explanation",
		"title",
		"causal_chains",
	},
	AdditionalProperties: false,
	Schema:               schema.URL,
}

type Relationship struct {
	From              string `json:"from"`
	To                string `json:"to"`
	Polarity          string `json:"polarity"` // "+", or "-"
	Reasoning         string `json:"reasoning"`
	PolarityReasoning string `json:"polarityReasoning"`
}

type RelationshipEntry struct {
	Variable          string `json:"variable"`
	Polarity          string `json:"polarity"` // "+", or "-"
	PolarityReasoning string `json:"polarityReasoning"`
}

type Chain struct {
	InitialVariable string              `json:"initial_variable"`
	Relationships   []RelationshipEntry `json:"relationships"`
	Reasoning       string              `json:"reasoning"`
}

type Map struct {
	Title        string  `json:"title"`
	Explanation  string  `json:"explanation"`
	CausalChains []Chain `json:"causal_chains"`
}

func (m *Map) Variables() (vars Set[string]) {
	vars = make(Set[string])
	for _, c := range m.CausalChains {
		vars.Add(c.InitialVariable)
		for _, next := range c.Relationships {
			vars.Add(next.Variable)
		}
	}
	return vars
}

type searchState struct {
	edges   map[string][]string
	visited Set[string]
	found   [][]string
}

func (s *searchState) addCycle(path []string) {
	cycle := make([]string, 0, len(path))

	// rotate the path so that the lowest-named variable is first
	i := slices.Index(path, slices.Min(path))
	cycle = append(cycle, path[i:]...)
	cycle = append(cycle, path[:i]...)

	for _, foundCycle := range s.found {
		// already recorded it, nothing to do
		if slices.Equal(foundCycle, cycle) {
			return
		}
	}

	s.found = append(s.found, cycle)
}

func (s *searchState) search(path []string, v string) {
	s.visited.Add(v)
	path = append(path, v)

	for _, neighbor := range s.edges[v] {
		if !s.visited.Contains(neighbor) {
			s.search(path, neighbor)
		}
		// found a cycle
		if i := slices.Index(path, neighbor); i >= 0 {
			s.addCycle(path[i:])
		}
	}
}

func findCycles(outgoing map[string][]string) (found [][]string) {
	s := searchState{
		edges:   outgoing,
		visited: make(Set[string], len(outgoing)),
	}

	for v := range outgoing {
		clear(s.visited)

		path := make([]string, 0, 32)
		s.search(path, v)
	}

	return s.found
}

func (m *Map) Loops() [][]string {
	// build a map of all outgoing edges in our diagram/graph.
	outgoing := make(map[string][]string)
	for _, chain := range m.CausalChains {
		for i, r := range chain.Relationships {
			var from string
			if i == 0 {
				from = chain.InitialVariable
			} else {
				from = chain.Relationships[i-1].Variable
			}
			outgoing[from] = append(outgoing[from], r.Variable)
		}
	}

	allLoops := findCycles(outgoing)

	// make the loops clearer by ensuring that we repeat as the last
	// element the initial one.
	for i, loop := range allLoops {
		allLoops[i] = append(loop, loop[0])
	}

	return allLoops
}

func (m *Map) VisualSVG() ([]byte, error) {
	var b strings.Builder

	b.WriteString("digraph {\n\toverlap=false\n\tmode=KK\n")

	// FIXME
	//for _, r := range m.Relationships {
	//	b.WriteString(fmt.Sprintf("\t%q -> %q\n", r.From, r.To))
	//}

	b.WriteString("}\n")

	cmd := exec.Command("dot", "-Tsvg", "-Ksfdp")
	cmd.Stdin = strings.NewReader(b.String())
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("cmd.StdoutPipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("cmd.Start: %w", err)
	}

	svg, err := io.ReadAll(stdout)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w", err)
	}

	if err = cmd.Wait(); err != nil {
		return nil, fmt.Errorf("cmd.Wait: %w ()", err)
	}

	return svg, nil
}
