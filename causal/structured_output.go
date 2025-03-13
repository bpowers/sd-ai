package causal

import (
	"cmp"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/isee-systems/sd-ai/schema"
)

//go:embed response_schema.json
var responseSchemaJson string

type Polarity int

const (
	NegativePolarity Polarity = iota
	PositivePolarity
)

func (p Polarity) IsPositive() bool {
	return p == PositivePolarity
}

func (p Polarity) IsNegative() bool {
	return !p.IsPositive()
}

func (p Polarity) Symbol() string {
	switch p {
	case PositivePolarity:
		return "+"
	default:
		return "-"
	}
}

func (p Polarity) String() string {
	return p.Symbol()
}

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

var RelationshipsResponseSchema *schema.JSON

func init() {
	RelationshipsResponseSchema = new(schema.JSON)
	err := json.Unmarshal([]byte(responseSchemaJson), RelationshipsResponseSchema)
	if err != nil {
		panic(err)
	}
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
		vars.Add(strings.TrimSpace(strings.ToLower(c.InitialVariable)))
		for _, next := range c.Relationships {
			vars.Add(strings.TrimSpace(strings.ToLower(next.Variable)))
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
			from = strings.TrimSpace(strings.ToLower(from))
			to := strings.TrimSpace(strings.ToLower(r.Variable))
			outgoing[from] = append(outgoing[from], to)
		}
	}

	allLoops := findCycles(outgoing)

	// make the loops clearer by ensuring that we repeat as the last
	// element the initial one.
	for i, loop := range allLoops {
		allLoops[i] = append(loop, loop[0])
	}

	slices.SortStableFunc(allLoops, func(a, b []string) int {
		if len(a) < len(b) {
			return -1
		} else if len(a) > len(b) {
			return 1
		}

		return slices.Compare(a, b)
	})

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

func NewMap(relationships []Relationship) *Map {
	m := &Map{}

	for _, r := range relationships {
		m.CausalChains = append(m.CausalChains, Chain{
			InitialVariable: r.From,
			Relationships: []RelationshipEntry{
				{
					Variable:          r.To,
					Polarity:          r.Polarity,
					PolarityReasoning: r.PolarityReasoning,
				},
			},
		})
	}

	return m
}
