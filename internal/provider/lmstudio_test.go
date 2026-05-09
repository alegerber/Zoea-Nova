package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
