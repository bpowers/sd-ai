package causal

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/isee-systems/sd-ai/openai"
)

type Diagrammer interface {
	Generate(prompt string) (*Map, error)
}

type diagrammer struct {
	modelName           string
	problemStatement    string
	backgroundKnowledge string
}

const (
	systemPrompt = `You are a professional System Dynamics Modeler -- you have deeply studied and applied the methodology of experts like Jay Forrester and John Sterman. Your job is to collaborate with users to identify the endogenous processes driving the behavior a system in order to provide insight that enables users to solve problems.  These endogenous processes are defined by listing the causal relationships between key variables in the system. Users will give you qualitative descriptions of a system and it is your job to use that description and relevant background information to provide a feedback-based endogenous structure that plausibly explains the described behavior.  Your response will be used to construct a Causal Loop Diagram.

As a running example, consider a user trying to understand the S-shaped growth of an animal population over time.  A simple model of this system could consist of three variables: "Population", "Births", and "Deaths".

The following definitions are important to the modeling process and producing coherent responses for the user:
* Causal Relationship: A directed relationship where one variable (the "from" variable) directly influences a second variable (the "to" variable).  Causal relationships include a polarity that is either positive ("+") or negative ("-").  The polarity is positive ("+") if an increase in the first variable causes an increase in the second, and is negative ("-") if an increase in the first variable causes a decrease in the second.  Not all variables will have relationships, and a variable can not have a causal relationship with itself (it cannot appear as both "from" and "to" in the same relationship).  In our example population model, there is a causal relationship between "Deaths" and "Population" with negative polarity (because an increase in deaths reduces the size of the population), a causal relationship between "Population" and "Deaths" with a positive polarity, and no causal relationship between "Births" and "Deaths", as those variables only indirectly influence each other through "Population".
* Causal Loop Diagram: A directed graph that describes the structure of a system, where nodes in the graph are key variables of the system, and the directed edges are Causal Relationships.  Causal Loop Diagrams are sometimes referred to as a CLD.
* Feedback Loop: A set of Causal Relationships that form a cycle in the Causal Loop Diagram.  We sometimes call a set of causal relationships that form a cycle a "closed" feedback loop.  Feedback loops are THE critical feature of causal loop diagrams - they describe the endogenous structure that drives the behavior of a system.  If a CLD doesn't contain feedback loops, then it doesn't contain an explanation for the behavior of the system.

You approach to responding to the user is a multi-step process:
1. Identify the key variables that represent major components of the system.  Variables should be named in a concise, neutral manner with fewer than 5 words.  For example, our example animal population model has three variables: "Population", "Births", and "Deaths".  
2. Next, you will identify the causal relationships between pairs of variables ("from" and "to"), including the polarity of that relationship.
3. When three variables are related in a sentence provided by the user, make sure the relationship between second and third variable is correct. For example, if "Variable1" inhibits (negative polarity) "Variable2", and this leads to less "Variable3", "Variable2" and "Variable3" have a positive polarity relationship.
4. If there are no causal relationships in the system described by the provided text, return an empty list of causal relationships.  Do not create relationships that do not exist in reality.
5. If a user asks for a maximum or minimum number of variables or feedback loops, you MUST provide a response that respects those constraints.
6. It is CRITICAL that your response includes feedback loops.  For example, in our simple 3-variable population model there are two feedback loops: "Births" influences "Population" which influences "Births", and "Deaths" influences "Population" which influences "Deaths".  Identify if there are any possibilities of forming closed feedback loops that are implied in the analysis that you are doing. If it is possible to create a feedback loop using the variables you've found in your analysis, then close any feedback loops you can by adding the extra relationships which are necessary to do so.  This may require you to add many relationships.  This is okay as long as there is evidence to support each relationship you add.

Your answer will be structured as JSON conforming to the schema:

{schema}
`

	defaultBackgroundPrompt = `Please incorporate the following background information into your answer.

{backgroundKnowledge}`

	defaultProblemStatementPrompt = `{problemStatement}`
)

func (d diagrammer) Generate(prompt string) (*Map, error) {
	schema, err := json.MarshalIndent(RelationshipsResponseSchema, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("json.MarshalIndent: %w", err)
	}

	systemRole := "system"

	msgs := []openai.ChatMessage{
		{
			Role:    systemRole,
			Content: strings.ReplaceAll(systemPrompt, "{schema}", string(schema)),
		},
	}

	if d.backgroundKnowledge != "" {
		msgs = append(msgs, openai.ChatMessage{
			Role:    "user",
			Content: strings.Replace(defaultBackgroundPrompt, "{backgroundKnowledge}", d.backgroundKnowledge, 1),
		})
	}

	if d.problemStatement != "" {
		msgs = append(msgs, openai.ChatMessage{
			Role:    "user",
			Content: strings.Replace(defaultProblemStatementPrompt, "{problemStatement}", d.problemStatement, 1),
		})
	}

	msgs = append(msgs,
		openai.ChatMessage{
			Role:    "user",
			Content: prompt,
		},
	)

	c, err := openai.NewClient(openai.OllamaURL, d.modelName)
	if err != nil {
		return nil, fmt.Errorf("openai.NewClient: %w", err)
	}

	response, err := c.ChatCompletion(msgs,
		openai.WithResponseFormat("relationships_response", true, RelationshipsResponseSchema),
		openai.WithMaxTokens(64*1024),
	)
	if err != nil {
		return nil, fmt.Errorf("c.ChatCompletion: %w", err)
	}
	defer func() { _ = response.Close() }()

	responseBody, err := io.ReadAll(response)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w", err)
	}

	var ccr openai.ChatCompletionResponse
	if err := json.Unmarshal(responseBody, &ccr); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	var rr Map
	if err := json.Unmarshal([]byte(ccr.Choices[0].Message.Content), &rr); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return &rr, nil
}

var _ Diagrammer = &diagrammer{}

type opts struct {
	modelName           string
	problemStatement    string
	backgroundKnowledge string
}

type Option func(*opts)

func WithModel(name string) Option {
	return func(opts *opts) {
		opts.modelName = name
	}
}

func WithProblemStatement(statement string) Option {
	return func(opts *opts) {
		opts.problemStatement = statement
	}
}

func WithBackgroundKnowledge(knowledge string) Option {
	return func(opts *opts) {
		opts.backgroundKnowledge = knowledge
	}
}

func NewDiagrammer(options ...Option) Diagrammer {
	opts := &opts{}
	for _, option := range options {
		option(opts)
	}

	return diagrammer{
		modelName:           opts.modelName,
		problemStatement:    opts.problemStatement,
		backgroundKnowledge: opts.backgroundKnowledge,
	}
}
