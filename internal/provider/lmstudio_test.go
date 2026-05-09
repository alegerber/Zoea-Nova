package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLMStudio_ChatBasic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path = %q, want /v1/chat/completions", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "" {
			t.Errorf("LM Studio should not send Authorization, got %q", r.Header.Get("Authorization"))
		}

		var req struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if req.Model != "qwen2.5-7b" {
			t.Errorf("model = %q, want qwen2.5-7b", req.Model)
		}
		if req.Messages[0].Role != "system" {
			t.Errorf("first message role = %q, want system", req.Messages[0].Role)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"role": "assistant", "content": "ok"}},
			},
		})
	}))
	defer server.Close()

	baseURL := strings.TrimSuffix(server.URL, "/v1")
	p := NewLMStudio(baseURL, "qwen2.5-7b")

	resp, err := p.Chat(context.Background(), []Message{
		{Role: "system", Content: "be concise"},
		{Role: "user", Content: "hi"},
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if resp != "ok" {
		t.Errorf("response = %q, want ok", resp)
	}
}

func TestLMStudio_EndpointAcceptsTrailingV1(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"role": "assistant", "content": "x"}},
			},
		})
	}))
	defer server.Close()

	p := NewLMStudio(server.URL+"/v1", "model-x")
	if _, err := p.Chat(context.Background(), []Message{{Role: "user", Content: "hi"}}); err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if hits != 1 {
		t.Errorf("hits = %d, want 1", hits)
	}
}

func TestLMStudio_Name(t *testing.T) {
	p := NewLMStudioWithTemp("lm-qwen", "http://localhost:1234", "qwen2.5", 0.7)
	if p.Name() != "lm-qwen" {
		t.Errorf("Name() = %q, want lm-qwen", p.Name())
	}
}

// TestLMStudio_RetryOn500 verifies the retry loop recovers from a transient 500 error.
func TestLMStudio_RetryOn500(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits == 1 {
			http.Error(w, "transient", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"role": "assistant", "content": "recovered"}},
			},
		})
	}))
	defer server.Close()

	// Use a no-delay variant: override the package-level retry delays for fast tests.
	saved := lmstudioRetryDelays
	lmstudioRetryDelays = []time.Duration{0, 0, 0}
	defer func() { lmstudioRetryDelays = saved }()

	p := NewLMStudio(server.URL, "test-model")
	resp, err := p.Chat(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if resp != "recovered" {
		t.Errorf("response = %q, want recovered", resp)
	}
	if hits != 2 {
		t.Errorf("hits = %d, want 2 (one fail + one retry)", hits)
	}
}

// TestLMStudio_NoRetryOn501 verifies that 501 Not Implemented is NOT retried
// (it's a permanent error, not a transient one).
func TestLMStudio_NoRetryOn501(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		http.Error(w, "not implemented", http.StatusNotImplemented)
	}))
	defer server.Close()

	saved := lmstudioRetryDelays
	lmstudioRetryDelays = []time.Duration{0, 0, 0}
	defer func() { lmstudioRetryDelays = saved }()

	p := NewLMStudio(server.URL, "test-model")
	_, err := p.Chat(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected error from 501, got nil")
	}
	if hits != 1 {
		t.Errorf("hits = %d, want 1 (no retry on 501)", hits)
	}
}

func TestLMStudio_ChatWithTools(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Tools []struct {
				Function struct {
					Name string `json:"name"`
				} `json:"function"`
			} `json:"tools"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(req.Tools) != 1 || req.Tools[0].Function.Name != "get_weather" {
			t.Errorf("expected one tool 'get_weather', got %+v", req.Tools)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{
				"message": map[string]any{
					"role":    "assistant",
					"content": "",
					"tool_calls": []map[string]any{{
						"id":   "call_1",
						"type": "function",
						"function": map[string]any{
							"name":      "get_weather",
							"arguments": `{"city":"berlin"}`,
						},
					}},
				},
			}},
		})
	}))
	defer server.Close()

	p := NewLMStudio(server.URL, "qwen2.5-7b")
	resp, err := p.ChatWithTools(context.Background(),
		[]Message{{Role: "user", Content: "weather"}},
		[]Tool{{
			Name:        "get_weather",
			Description: "Get weather",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}}}`),
		}},
	)
	if err != nil {
		t.Fatalf("ChatWithTools: %v", err)
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(resp.ToolCalls))
	}
	if resp.ToolCalls[0].Name != "get_weather" {
		t.Errorf("tool call name = %q", resp.ToolCalls[0].Name)
	}
}
