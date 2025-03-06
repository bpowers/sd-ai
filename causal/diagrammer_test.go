package causal

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

var testMap1 = &Map{
	Title:       "American Revolution Onset",
	Explanation: "Based on historical context and user input,",
	Relationships: []Relationship{
		{
			From:              "Tax Burden",
			To:                "Tensions",
			Polarity:          "+",
			Reasoning:         "The British government imposed various taxes, such as the Stamp Act and Townshend Acts, which increased the financial burden on American colonies.",
			PolarityReasoning: "An increase in Tax Burden led to an increase in Tensions.",
		},
		{
			From:              "Tax Burden",
			To:                "Resistance",
			Polarity:          "+",
			Reasoning:         "High taxes fueled protests and boycotts against British goods, demonstrating growing resistance among colonists.",
			PolarityReasoning: "An increase in Tax Burden led to an increase in Resistance.",
		},
		{
			From:              "Tensions",
			To:                "Clashes",
			Polarity:          "+",
			Reasoning:         "Escalating tensions between British authorities and American patriots raised the probability of violent confrontations.",
			PolarityReasoning: "Rising Tensions increased the likelihood of Clashes.",
		},
		{
			From:              "Resistance",
			To:                "Clashes",
			Polarity:          "+",
			Reasoning:         "Increased resistance through protests, boycotts, and other forms of dissent heightened the risk of physical confrontations with British forces.",
			PolarityReasoning: "As Resistance grew, so did the likelihood of Clashes.",
		},
		{
			From:              "Clashes",
			To:                "Tensions",
			Polarity:          "+",
			Reasoning:         "Violent encounters between colonists and British troops intensified feelings of hostility and mistrust, fueling a cycle of escalating violence.",
			PolarityReasoning: "An increase in Clashes increased Tensions further.",
		},
		{
			From:              "Clashes",
			To:                "Resistance",
			Polarity:          "+",
			Reasoning:         "Each clash between the British and the colonists served to galvanize support among the population for independence, strengthening the resolve of those resisting British authority.",
			PolarityReasoning: "An increase in Clashes also increased Resistance as more colonists became determined to fight against British rule.",
		},
		{
			From:              "Tensions",
			To:                "Tax Burden",
			Polarity:          "+",
			Reasoning:         "As tensions rose, the British government responded with stricter enforcement of its authority and additional taxation measures, aiming to quell dissent and maintain control.",
			PolarityReasoning: "Increased Tensions led to increased Tax Burden as Britain attempted to assert its control over the colonies more firmly.",
		},
	},
}

func TestExtractingResults(t *testing.T) {
	causalMap := testMap1

	vars := causalMap.Variables()
	expectedVars := NewSet(
		"Tax Burden",
		"Resistance",
		"Clashes",
		"Tensions",
	)
	assert.Equal(t, expectedVars, vars)

	loops := causalMap.Loops()
	assert.Contains(t, loops, []string{"Clashes", "Tensions", "Clashes"})
	assert.Contains(t, loops, []string{"Clashes", "Resistance", "Clashes"})
	assert.Contains(t, loops, []string{"Tax Burden", "Tensions", "Tax Burden"})
	assert.Contains(t, loops, []string{"Clashes", "Tensions", "Tax Burden", "Resistance", "Clashes"})
	assert.Equal(t, 4, len(loops))
}

func TestDiagrammerSVG(t *testing.T) {
	svg, err := testMap1.VisualSVG()
	require.NoError(t, err)

	// assert we got something
	assert.Greater(t, len(svg), 0)

	//f, err := os.CreateTemp("", "cld-*.svg")
	//require.NoError(t, err)
	//
	//n, err := f.Write(svg)
	//require.NoError(t, err)
	//require.Equal(t, len(svg), n)

	//path := f.Name()
	//require.NoError(t, f.Close())
	//
	//err = exec.Command("open", path).Run()
	//require.NoError(t, err)
}
