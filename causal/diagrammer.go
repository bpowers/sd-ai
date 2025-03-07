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
	defaultSystemPrompt = `You are a System Dynamics Professional Modeler. Users will give you text, and it is your job to generate causal relationships from that text.

You will conduct a multistep process:

1. You will identify all the entities that have a cause-and-effect relationships between. These entities are variables. Name these variables in a concise manner. A variable name should not be more than 5 words. Make sure that you minimize the number of variables used. Variable names should be neutral, i.e., there shouldn't be positive or negative meaning in variable names.

2. For each variable, represent its causal relationships with other variables. There are two different kinds of polarities for causal relationships: positive polarity represented with a + symbol and negative represented with a - symbol. A positive polarity (+) relationship exits when variables are positively correlated.  Here are two examples of positive polarity (+) relationships. If a decline in the causing variable (the from variable) leads to a decline in the effect variable (the to variable), then the relationship has a positive polarity (+).  A relationship also has a positive polarity (+) if an increase in the causing variable (the from variable) leads to an increase in the effect variable (the to variable).  A negative polarity (-) is when variables are anticorrelated.  Here are two examples of negative polarity (-) relationships.  If a decline in the causing variable (the from variable) leads to an increase in the effect variable (the to variable), then the relationship has a negative polarity (-). A relationship also has a negative polarity (-) if an increase in the causing variable (the from variable) causes a decrease in the effect variable (the to variable). 

3. Not all variables will have relationships with all other variables.

4. When three variables are related in a sentence, make sure the relationship between second and third variable is correct. For example, if "Variable1" inhibits "Variable2", leading to less "Variable3", "Variable2" and "Variable3" have a positive polarity (+) relationship.

5. If there are no causal relationships at all in the provided text, return an empty JSON array.  Do not create relationships which do not exist in reality.

6. Try as hard as you can to close feedback loops between the variables you find. It is very important that your answer includes feedback.  A feedback loop happens when there is a closed causal chain of relationships.  An example would be “Variable1” causes “Variable2” to increase, which causes “Variable3” to decrease which causes “Variable1” to again increase.  Try to find as many of the feedback loops as you can.`

	defaultBackgroundPrompt = `Please be sure to consider the following critically important background information when you give your answer.

{backgroundKnowledge}`

	defaultFeedbackPrompt = `Find out if there are any possibilities of forming closed feedback loops that are implied in the analysis that you are doing. If it is possible to create a feedback loop using the variables you've found in your analysis, then close any feedback loops you can by adding the extra relationships which are necessary to do so.  This may require you to add many relationships.  This is okay as long as there is evidence to support each relationship you add.`

	defaultProblemStatementPrompt = `The user has stated that they are conducting this modeling exercise to understand the following problem better.

{problemStatement}`
)

func (d diagrammer) Generate(prompt string) (*Map, error) {
	systemRole := "system"

	msgs := []openai.ChatMessage{
		{
			Role:    systemRole,
			Content: defaultSystemPrompt,
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
			Role:    systemRole,
			Content: strings.Replace(defaultProblemStatementPrompt, "{problemStatement}", d.problemStatement, 1),
		})
	}

	msgs = append(msgs,
		openai.ChatMessage{
			Role:    "user",
			Content: prompt,
		},
		openai.ChatMessage{
			Role:    "user",
			Content: defaultFeedbackPrompt,
		},
	)

	c, err := openai.NewClient(openai.OllamaURL, d.modelName)
	if err != nil {
		return nil, fmt.Errorf("openai.NewClient: %w", err)
	}

	response, err := c.ChatCompletion(msgs, openai.WithResponseFormat("relationships_response", true, RelationshipsResponseSchema))
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
