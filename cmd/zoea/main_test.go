package main

import (
	"testing"

	"github.com/xonecas/zoea-nova/internal/config"
)

// TestInitProviders_TypeFieldRouting verifies type="ollama" routes to the Ollama factory.
func TestInitProviders_TypeFieldRouting(t *testing.T) {
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"explicit-ollama": {
				Type:        "ollama",
				Endpoint:    "http://example.invalid:9999",
				Model:       "qwen3:8b",
				Temperature: 0.7,
			},
		},
	}
	registry := initProviders(cfg, &config.Credentials{})
	if _, err := registry.Create("explicit-ollama", "qwen3:8b", 0.7); err != nil {
		t.Errorf("explicit ollama type should register: %v", err)
	}
}

// TestInitProviders_LegacyFallback verifies endpoint heuristic still works when type is empty.
func TestInitProviders_LegacyFallback(t *testing.T) {
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"ollama-qwen": {
				Endpoint:    "http://localhost:11434",
				Model:       "qwen3:8b",
				Temperature: 0.7,
			},
		},
	}
	registry := initProviders(cfg, &config.Credentials{})
	if _, err := registry.Create("ollama-qwen", "qwen3:8b", 0.7); err != nil {
		t.Errorf("legacy ollama-qwen should still register, got %v", err)
	}
}
