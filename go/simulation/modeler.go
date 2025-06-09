package simulation

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/UB-IAD/sd-ai/go/chat"
)

type Modeler interface {
	Generate(ctx context.Context, prompt, backgroundKnowledge string) (*Model, error)
}

type modeler struct {
	client chat.Client
}

var _ Modeler = &modeler{}

func NewModeler(client chat.Client) Modeler {
	return modeler{
		client: client,
	}
}

var (
	//go:embed system_prompt.txt
	baseSystemPrompt string

	//go:embed background_prompt.txt
	backgroundPrompt string
)

func (d modeler) Generate(ctx context.Context, prompt, backgroundKnowledge string) (*Model, error) {
	schema, err := json.MarshalIndent(RelationshipsResponseSchema, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("json.MarshalIndent: %w", err)
	}

	systemPrompt := strings.ReplaceAll(baseSystemPrompt, "{schema}", string(schema))

	msg := chat.Message{
		Role: chat.UserRole,
		Content: fmt.Sprintf("%s\n\n%s",
			strings.ReplaceAll(backgroundPrompt, "{backgroundKnowledge}", backgroundKnowledge),
			prompt,
		),
	}

	c := d.client.NewChat(systemPrompt)

	resp, err := c.Message(ctx, msg,
		chat.WithResponseFormat("relationships_response", true, RelationshipsResponseSchema),
		chat.WithMaxTokens(64*1024),
	)
	if err != nil {
		return nil, fmt.Errorf("c.ChatCompletion: %w", err)
	}

	var rr Model
	if err := json.Unmarshal([]byte(resp.Content), &rr); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return &rr, nil
}
