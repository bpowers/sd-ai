package causal

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/isee-systems/sd-ai/chat"
	"github.com/isee-systems/sd-ai/openai"
)

type Diagrammer interface {
	Generate(ctx context.Context, prompt, backgroundKnowledge string) (*Map, error)
}

type diagrammer struct {
	client chat.Client
}

var (
	//go:embed system_prompt.txt
	systemPrompt string

	//go:embed background_prompt.txt
	backgroundPrompt string
)

func (d diagrammer) Generate(ctx context.Context, prompt, backgroundKnowledge string) (*Map, error) {
	schema, err := json.MarshalIndent(RelationshipsResponseSchema, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("json.MarshalIndent: %w", err)
	}

	var msgs []chat.Message

	if backgroundKnowledge != "" {
		msgs = append(msgs, chat.Message{
			Role:    chat.UserRole,
			Content: strings.ReplaceAll(backgroundPrompt, "{backgroundKnowledge}", backgroundKnowledge),
		})
	}

	msgs = append(msgs, chat.Message{
		Role:    chat.UserRole,
		Content: prompt,
	})

	response, err := d.client.ChatCompletion(ctx, msgs,
		chat.WithResponseFormat("relationships_response", true, RelationshipsResponseSchema),
		chat.WithMaxTokens(64*1024),
		chat.WithSystemPrompt(strings.ReplaceAll(systemPrompt, "{schema}", string(schema))),
	)
	if err != nil {
		return nil, fmt.Errorf("c.ChatCompletion: %w", err)
	}

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

func NewDiagrammer(client chat.Client) Diagrammer {
	return diagrammer{
		client: client,
	}
}
