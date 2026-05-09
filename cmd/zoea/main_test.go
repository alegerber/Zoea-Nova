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

// TestInitProviders_OpenCodeMissingKeySkips verifies that an opencode-typed provider
// without an API key is skipped (and the bug from earlier sessions where this caused
// "provider not found" with no log hint is now logged at warn level).
func TestInitProviders_OpenCodeMissingKeySkips(t *testing.T) {
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"zen-pickle": {
				Type:        "opencode",
				Endpoint:    "https://opencode.ai/zen/v1",
				Model:       "big-pickle",
				APIKeyName:  "opencode_zen",
				Temperature: 0.7,
			},
		},
	}
	registry := initProviders(cfg, &config.Credentials{})

	if _, err := registry.Create("zen-pickle", "big-pickle", 0.7); err == nil {
		t.Error("expected provider to be unregistered when API key is missing")
	}
}

// TestInitProviders_UnknownTypeSkips verifies the default branch — adapterType empty
// because the endpoint matches no heuristic and Type is unset. This locks in the
// fallback behaviour before Task 8 adds the lmstudio branch.
func TestInitProviders_UnknownTypeSkips(t *testing.T) {
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"mystery": {
				Endpoint:    "https://unknown-vendor.example/v1",
				Model:       "mystery-7b",
				Temperature: 0.7,
			},
		},
	}
	registry := initProviders(cfg, &config.Credentials{})

	if _, err := registry.Create("mystery", "mystery-7b", 0.7); err == nil {
		t.Error("expected unknown-type provider to be unregistered")
	}
}

func TestInitProviders_LMStudioByType(t *testing.T) {
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"lm-qwen": {
				Type:        "lmstudio",
				Endpoint:    "http://localhost:1234/v1",
				Model:       "qwen2.5-7b",
				Temperature: 0.5,
			},
		},
	}
	registry := initProviders(cfg, &config.Credentials{})

	p, err := registry.Create("lm-qwen", "qwen2.5-7b", 0.5)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if p.Name() != "lm-qwen" {
		t.Errorf("Name() = %q, want lm-qwen", p.Name())
	}
}
