package chat

import (
	"context"
	"io"

	"github.com/isee-systems/sd-ai/schema"
)

type requestOpts struct {
	temperature     *float64
	reasoningEffort string
	responseFormat  *JsonSchema
	maxTokens       int
	systemPrompt    string
}

type Options struct {
	Temperature     *float64
	ReasoningEffort string
	ResponseFormat  *JsonSchema
	MaxTokens       int
	SystemPrompt    string
}

type JsonSchema struct {
	Name   string       `json:"name"`
	Strict bool         `json:"strict,omitempty"`
	Schema *schema.JSON `json:"schema,omitempty"`
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
		opts.responseFormat = &JsonSchema{
			Name:   name,
			Strict: strict,
			Schema: schema,
		}
	}
}

func WithSystemPrompt(prompt string) Option {
	return func(opts *requestOpts) {
		opts.systemPrompt = prompt
	}
}

func ApplyOptions(opts ...Option) Options {
	var options requestOpts
	for _, opt := range opts {
		opt(&options)
	}

	return Options{
		Temperature:     options.temperature,
		ReasoningEffort: options.reasoningEffort,
		ResponseFormat:  options.responseFormat,
		MaxTokens:       options.maxTokens,
		SystemPrompt:    options.systemPrompt,
	}
}

const (
	UserRole   = "user"
	SystemRole = "system"
)

type Client interface {
	ChatCompletion(ctx context.Context, msgs []Message, opts ...Option) (io.Reader, error)
}

type Message struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type debugDirContextKey struct{}

func WithDebugDir(ctx context.Context, dir string) context.Context {
	return context.WithValue(ctx, debugDirContextKey{}, dir)
}

func DebugDir(ctx context.Context) string {
	return ctx.Value(debugDirContextKey{}).(string)
}
