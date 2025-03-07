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
		"explanation": {
			Type:        schema.String,
			Description: "Concisely explain your reasoning for each change you made to the old CLD to create the new CLD. Speak in plain English, don't reference json specifically. Don't reiterate the request or any of these instructions.",
		},
		"title": {
			Type:        schema.String,
			Description: "A highly descriptive 7 word max title describing your explanation.",
		},
		"relationships": {
			Type: schema.Array,
			Items: &schema.JSON{
				Type: schema.Object,
				Properties: map[string]*schema.JSON{
					"from": {
						Type:        schema.String,
						Description: "This is a variable which causes the to variable in this relationship that is between two variables, from and to.  The from variable is the equivalent of a cause.  The to variable is the equivalent of an effect",
					},
					"to": {
						Type:        schema.String,
						Description: "This is a variable which is impacted by the from variable in this relationship that is between two variables, from and to.  The from variable is the equivalent of a cause.  The to variable is the equivalent of an effect",
					},
					"polarity": {
						Type: schema.String,
						Enum: []string{
							"+",
							"-",
						},
						Description: "There are two possible kinds of relationships.  The first are relationships with positive polarity that are represented with a + symbol.  In relationships with positive polarity (+) a change in the from variable causes a change in the same direction in the to variable.  For example, in a relationship with postive polarity (+), a decrease in the from variable, would lead to a decrease in the to variable.  The second kind of relationship are those with negative polarity that are represented with a - symbol.  In relationships with negative polarity (-) a change in the from variable causes a change in the opposite direction in the to variable.  For example, in a relationship with negative polarity (-) an increase in the from variable, would lead to a decrease in the to variable.",
					},
					"reasoning": {
						Type:        schema.String,
						Description: "This is an explanation for why this relationship exists",
					},
					"polarityReasoning": {
						Type:        schema.String,
						Description: "This is the reason for why the polarity for this relationship was choosen",
					},
				},
				Required: []string{
					"from",
					"to",
					"polarity",
					"reasoning",
					"polarityReasoning",
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
		"relationships",
	},
	AdditionalProperties: false,
	Schema:               "http://json-schema.org/draft-07/schema#",
}

type Relationship struct {
	From              string `json:"from"`
	To                string `json:"to"`
	Polarity          string `json:"polarity"` // "+", or "-"
	Reasoning         string `json:"reasoning"`
	PolarityReasoning string `json:"polarityReasoning"`
}

type Map struct {
	Title         string         `json:"title"`
	Explanation   string         `json:"explanation"`
	Relationships []Relationship `json:"relationships"`
}

func (m *Map) Variables() (vars Set[string]) {
	vars = make(Set[string])
	for _, r := range m.Relationships {
		vars.Add(r.From)
		vars.Add(r.To)
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
	for _, r := range m.Relationships {
		outgoing[r.From] = append(outgoing[r.From], r.To)
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

	for _, r := range m.Relationships {
		b.WriteString(fmt.Sprintf("\t%q -> %q\n", r.From, r.To))
	}

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
