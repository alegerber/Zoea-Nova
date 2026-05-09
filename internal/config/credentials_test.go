package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCredentials_RegistrationCodeRoundtrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	creds := &Credentials{
		RegistrationCode: "ABC-123",
		Providers: map[string]ProviderCredentials{
			"opencode_zen": {APIKey: "sk-test"},
		},
	}
	if err := SaveCredentials(creds); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}

	loaded, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials: %v", err)
	}
	if loaded.RegistrationCode != "ABC-123" {
		t.Errorf("RegistrationCode = %q, want ABC-123", loaded.RegistrationCode)
	}
	if loaded.Providers["opencode_zen"].APIKey != "sk-test" {
		t.Errorf("APIKey lost in roundtrip: %q", loaded.Providers["opencode_zen"].APIKey)
	}
}

func TestCredentials_RegistrationCodeOptionalLegacy(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	if err := os.MkdirAll(filepath.Join(dir, ".zoea-nova"), 0700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	legacy := `{
		"providers": {
			"opencode_zen": {"api_key": "sk-old"}
		}
	}`
	path := filepath.Join(dir, ".zoea-nova", "credentials.json")
	if err := os.WriteFile(path, []byte(legacy), 0600); err != nil {
		t.Fatalf("write legacy: %v", err)
	}

	loaded, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials: %v", err)
	}
	if loaded.RegistrationCode != "" {
		t.Errorf("legacy file should leave RegistrationCode empty, got %q", loaded.RegistrationCode)
	}
	if loaded.Providers["opencode_zen"].APIKey != "sk-old" {
		t.Errorf("APIKey not loaded: %q", loaded.Providers["opencode_zen"].APIKey)
	}
}
