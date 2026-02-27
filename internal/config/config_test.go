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
	if cfg.AI.Model != "claude-3-5-sonnet-20241022" {
		t.Errorf("expected model claude-3-5-sonnet-20241022, got %s", cfg.AI.Model)
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

func TestLoadConfigFromPath_FileNotFound(t *testing.T) {
	_, err := config.LoadConfigFromPath("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestLoadConfigFromPath_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(":::invalid yaml:::"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := config.LoadConfigFromPath(filepath.Join(dir, "config.yaml"))
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestLoadConfigFromPath_MissingAPIUser(t *testing.T) {
	dir := t.TempDir()
	content := "api_token: token\nworkspace_id: \"123\"\n"
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := config.LoadConfigFromPath(filepath.Join(dir, "config.yaml"))
	if err == nil {
		t.Error("expected error for missing api_user, got nil")
	}
}

func TestLoadConfigFromPath_MissingWorkspaceID(t *testing.T) {
	dir := t.TempDir()
	content := "api_user: user\napi_token: token\n"
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := config.LoadConfigFromPath(filepath.Join(dir, "config.yaml"))
	if err == nil {
		t.Error("expected error for missing workspace_id, got nil")
	}
}

func TestLoadConfigFromPath_MissingCredentials(t *testing.T) {
	dir := t.TempDir()
	content := "api_user: user\nworkspace_id: \"123\"\n"
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := config.LoadConfigFromPath(filepath.Join(dir, "config.yaml"))
	if err == nil {
		t.Error("expected error for missing credentials, got nil")
	}
}
