package openai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/isee-systems/sd-ai/schema"
)

const (
	OpenAIURL = "https://api.openai.com/v1"
	OllamaURL = "http://localhost:11434/v1"
)

type Client interface {
	ChatCompletion(msgs []ChatMessage, opts ...Option) (io.ReadCloser, error)
}

type client struct {
	apiBaseUrl string
	modelName  string
}

var _ Client = &client{}

type jsonSchema struct {
	Name   string       `json:"name"`
	Strict bool         `json:"strict,omitempty"`
	Schema *schema.JSON `json:"schema,omitempty"`
}

type requestOpts struct {
	temperature     *float64
	reasoningEffort string
	responseFormat  *jsonSchema
	maxTokens       int
}

type Option func(*requestOpts)

func WithTemperature(t float64) Option {
	return func(opts *requestOpts) {
		opts.temperature = &t
	}
}

func WithReasoningEffort(lowMedHigh string) Option {
	return func(opts *requestOpts) {
		opts.reasoningEffort = lowMedHigh
	}
}

func WithMaxTokens(tokens int) Option {
	return func(opts *requestOpts) {
		opts.maxTokens = tokens
	}
}

func WithResponseFormat(name string, strict bool, schema *schema.JSON) Option {
	return func(opts *requestOpts) {
		opts.responseFormat = &jsonSchema{
			Name:   name,
			Strict: strict,
			Schema: schema,
		}
	}
}

func NewClient(apiBase, modelName string) (Client, error) {
	return &client{
		apiBaseUrl: apiBase,
		modelName:  modelName,
	}, nil
}

type ChatMessage struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type responseFormat struct {
	Type       string      `json:"type"`
	JsonSchema *jsonSchema `json:"json_schema,omitempty"`
}

type chatCompletionRequest struct {
	Messages        []ChatMessage   `json:"messages"`
	Model           string          `json:"model,omitempty"`
	ResponseFormat  *responseFormat `json:"response_format,omitempty"`
	Temperature     *float64        `json:"temperature,omitempty"`
	ReasoningEffort string          `json:"reasoning_effort,omitempty"`
	MaxTokens       int             `json:"max_tokens,omitempty"`
}

func (c client) ChatCompletion(msgs []ChatMessage, opts ...Option) (io.ReadCloser, error) {
	reqOpts := &requestOpts{}
	for _, opt := range opts {
		opt(reqOpts)
	}

	req := &chatCompletionRequest{
		Messages:        msgs,
		Model:           c.modelName,
		Temperature:     reqOpts.temperature,
		ReasoningEffort: reqOpts.reasoningEffort,
	}

	if reqOpts.responseFormat != nil {
		req.ResponseFormat = &responseFormat{
			Type:       "json_schema",
			JsonSchema: reqOpts.responseFormat,
		}
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %w", err)
	}
	body := strings.NewReader(string(bodyBytes))

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

	return resp.Body, nil
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
