package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rex-chang/tapd-cli/internal/config"
)

func TestLoadConfig_WithAI(t *testing.T) {
	dir := t.TempDir()
	content := `
api_user: testuser
api_token: testtoken
workspace_id: "12345"
ai:
  provider: claude
  api_key: sk-test-key
  model: claude-3-5-sonnet-20241022
`
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadConfigFromPath(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatalf("LoadConfigFromPath failed: %v", err)
	}

	if cfg.AI.Provider != "claude" {
		t.Errorf("expected provider claude, got %s", cfg.AI.Provider)
	}
	if cfg.AI.APIKey != "sk-test-key" {
		t.Errorf("expected api_key sk-test-key, got %s", cfg.AI.APIKey)
	}
}

func TestLoadConfig_WithoutAI(t *testing.T) {
	dir := t.TempDir()
	content := `
api_user: testuser
api_token: testtoken
workspace_id: "12345"
`
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadConfigFromPath(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatalf("LoadConfigFromPath failed: %v", err)
	}

	if cfg.AI.Provider != "" {
		t.Errorf("expected empty provider, got %s", cfg.AI.Provider)
	}
}
