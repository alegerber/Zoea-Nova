package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	openai "github.com/sashabaranov/go-openai"
)

// LMStudioProvider implements the Provider interface for a local LM Studio server.
type LMStudioProvider struct {
	name        string
	client      *openai.Client
	baseURL     string
	httpClient  *http.Client
	model       string
	temperature float64
}

var lmstudioRetryDelays = []time.Duration{5 * time.Second, 10 * time.Second, 15 * time.Second}

// NewLMStudio creates a new LM Studio provider with default name and temperature.
func NewLMStudio(endpoint, model string) *LMStudioProvider {
	return NewLMStudioWithTemp("lmstudio", endpoint, model, 0.7)
}

// NewLMStudioWithTemp creates a new LM Studio provider with explicit name and temperature.
// The endpoint may or may not end in /v1; the constructor normalizes both forms.
func NewLMStudioWithTemp(name string, endpoint, model string, temperature float64) *LMStudioProvider {
	cfg := openai.DefaultConfig("")
	baseURL := strings.TrimRight(endpoint, "/")
	if !strings.HasSuffix(baseURL, "/v1") {
		baseURL = baseURL + "/v1"
	}
	cfg.BaseURL = baseURL

	return &LMStudioProvider{
		name:        name,
		client:      openai.NewClientWithConfig(cfg),
		baseURL:     baseURL,
		httpClient:  &http.Client{},
		model:       model,
		temperature: temperature,
	}
}

func (p *LMStudioProvider) Name() string { return p.name }

func (p *LMStudioProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	resp, err := p.createChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       p.model,
		Messages:    mergeSystemMessagesOpenAI(toOpenAIMessages(messages)),
		Temperature: float32(p.temperature),
		Stream:      false,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("no response choices")
	}
	return resp.Choices[0].Message.Content, nil
}

func (p *LMStudioProvider) Close() error {
	if p.httpClient != nil {
		p.httpClient.CloseIdleConnections()
	}
	return nil
}

// lmstudioRequest is a custom request struct to ensure stream:false is serialized
// The openai.ChatCompletionRequest has omitempty on Stream, which omits false values
type lmstudioRequest struct {
	Model       string                         `json:"model"`
	Messages    []openai.ChatCompletionMessage `json:"messages"`
	Tools       []openai.Tool                  `json:"tools,omitempty"`
	Temperature float32                        `json:"temperature,omitempty"`
	Stream      bool                           `json:"stream"` // NO omitempty - always serialize
}

func (p *LMStudioProvider) createChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (*openaiChatResponse, error) {
	customReq := lmstudioRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Tools:       req.Tools,
		Temperature: req.Temperature,
		Stream:      req.Stream,
	}
	body, err := json.Marshal(customReq)
	if err != nil {
		return nil, err
	}

	url := p.baseURL + "/chat/completions"
	maxRetries := len(lmstudioRetryDelays)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := lmstudioRetryDelays[attempt-1]
			log.Warn().
				Str("provider", p.name).
				Int("attempt", attempt).
				Dur("delay", delay).
				Msg("Retrying LM Studio request")
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		if attempt == 0 {
			log.Info().
				Str("provider", p.name).
				Str("model", req.Model).
				Int("message_count", len(req.Messages)).
				Int("tool_count", len(req.Tools)).
				Msg("LM Studio request started")
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := p.httpClient.Do(httpReq)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil, err
			}
			lastErr = err
			continue
		}

		if resp.StatusCode == 429 || resp.StatusCode == 500 || resp.StatusCode == 502 ||
			resp.StatusCode == 503 || resp.StatusCode == 504 {
			payload, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("chat completion status %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			payload, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("chat completion status %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("read response body: %w", err)
		}

		var decoded openaiChatResponse
		if err := json.Unmarshal(bodyBytes, &decoded); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}

		log.Info().
			Str("provider", p.name).
			Str("model", p.model).
			Int("attempt", attempt+1).
			Int("choice_count", len(decoded.Choices)).
			Msg("LM Studio request successful")

		return &decoded, nil
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
}

// ChatWithTools sends messages with available tools and returns response with potential tool calls.
func (p *LMStudioProvider) ChatWithTools(ctx context.Context, messages []Message, tools []Tool) (*ChatResponse, error) {
	openaiTools, err := toOpenAITools(tools)
	if err != nil {
		return nil, fmt.Errorf("invalid tool schema: %w", err)
	}

	resp, err := p.createChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       p.model,
		Messages:    mergeSystemMessagesOpenAI(toOpenAIMessages(messages)),
		Tools:       openaiTools,
		Temperature: float32(p.temperature),
		Stream:      false,
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		log.Error().
			Str("provider", p.name).
			Msg("LM Studio returned empty choices array")
		return nil, errors.New("no response choices")
	}

	choice := resp.Choices[0]
	result := &ChatResponse{
		Content:   choice.Message.Content,
		Reasoning: "", // OpenAI standard doesn't provide reasoning field
	}

	if len(choice.Message.ToolCalls) > 0 {
		result.ToolCalls = make([]ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			result.ToolCalls[i] = ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: json.RawMessage(tc.Function.Arguments),
			}
		}
	}

	return result, nil
}

// Stream — placeholder, full impl in Task 5.
func (p *LMStudioProvider) Stream(ctx context.Context, messages []Message) (<-chan StreamChunk, error) {
	return nil, errors.New("Stream not implemented yet")
}
