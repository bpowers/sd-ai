package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/isee-systems/sd-ai/chat"
)

const (
	OpenAIURL = "https://api.openai.com/v1"
	OllamaURL = "http://localhost:11434/v1"
)

type client struct {
	apiBaseUrl string
	modelName  string
}

var _ chat.Client = &client{}

func NewClient(apiBase, modelName string) (chat.Client, error) {
	return &client{
		apiBaseUrl: apiBase,
		modelName:  modelName,
	}, nil
}

type responseFormat struct {
	Type       string           `json:"type"`
	JsonSchema *chat.JsonSchema `json:"json_schema,omitempty"`
}

type chatCompletionRequest struct {
	Messages        []chat.Message  `json:"messages"`
	Model           string          `json:"model,omitempty"`
	ResponseFormat  *responseFormat `json:"response_format,omitempty"`
	Temperature     *float64        `json:"temperature,omitempty"`
	ReasoningEffort string          `json:"reasoning_effort,omitempty"`
	MaxTokens       int             `json:"max_tokens,omitempty"`
}

func (c client) ChatCompletion(ctx context.Context, msgs []chat.Message, opts ...chat.Option) (io.Reader, error) {
	reqOpts := chat.ApplyOptions(opts...)

	// for OpenAI models, the system prompt is the first message in the list of messages
	if reqOpts.SystemPrompt != "" {
		allMsgs := make([]chat.Message, 0, len(msgs)+1)
		allMsgs = append(allMsgs, chat.Message{
			Role:    chat.SystemRole,
			Content: reqOpts.SystemPrompt,
		})
		allMsgs = append(allMsgs, msgs...)
		msgs = allMsgs
	}

	req := &chatCompletionRequest{
		Messages:        msgs,
		Model:           c.modelName,
		Temperature:     reqOpts.Temperature,
		ReasoningEffort: reqOpts.ReasoningEffort,
	}

	if reqOpts.ResponseFormat != nil {
		req.ResponseFormat = &responseFormat{
			Type:       "json_schema",
			JsonSchema: reqOpts.ResponseFormat,
		}
	}

	bodyBytes, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %w", err)
	}
	body := strings.NewReader(string(bodyBytes))

	if debugDir := chat.DebugDir(ctx); debugDir != "" {
		outputPath := path.Join(debugDir, "request.json")
		if err = os.WriteFile(outputPath, bodyBytes, 0o644); err != nil {
			return nil, fmt.Errorf("os.WriteFile(%s): %w", outputPath, err)
		}
	}

	httpReq, err := http.NewRequest(http.MethodPost, c.apiBaseUrl+"/chat/completions", body)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http.DefaultClient.Do: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("http status code: %d (%s)", resp.StatusCode, string(body))
	}

	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll(resp.Body): %w", err)
	}

	if debugDir := chat.DebugDir(ctx); debugDir != "" {
		outputPath := path.Join(debugDir, "response.json")
		if err = os.WriteFile(outputPath, bodyBytes, 0o644); err != nil {
			return nil, fmt.Errorf("os.WriteFile(%s): %w", outputPath, err)
		}
	}

	return strings.NewReader(string(bodyBytes)), nil
}

type ChatCompletionChoice struct {
	Index   int `json:"index"`
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
}

type ChatCompletionResponse struct {
	Id      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int                    `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
}
