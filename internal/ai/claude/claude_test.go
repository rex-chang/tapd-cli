package claude_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rex-chang/tapd-cli/internal/ai"
	"github.com/rex-chang/tapd-cli/internal/ai/claude"
)

func TestClaudeProvider_Chat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") == "" {
			t.Error("missing x-api-key header")
		}
		if r.Header.Get("anthropic-version") == "" {
			t.Error("missing anthropic-version header")
		}
		resp := map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": "这是 AI 的回复"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := claude.NewProvider("test-key", "claude-3-5-sonnet-20241022",
		claude.WithBaseURL(server.URL))

	resp, err := p.Chat(context.Background(), []ai.Message{
		{Role: "user", Content: "你好"},
	})
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}
	if resp != "这是 AI 的回复" {
		t.Errorf("unexpected response: %s", resp)
	}
}

func TestClaudeProvider_Chat_WithSystemMessage(t *testing.T) {
	var capturedBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedBody)
		resp := map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": "ok"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := claude.NewProvider("key", "model", claude.WithBaseURL(server.URL))
	_, err := p.Chat(context.Background(), []ai.Message{
		{Role: "system", Content: "你是助手"},
		{Role: "user", Content: "你好"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// 验证 system message 被正确分离为顶层 system 字段
	if capturedBody["system"] != "你是助手" {
		t.Errorf("expected system prompt '你是助手', got %v", capturedBody["system"])
	}
	messages, _ := capturedBody["messages"].([]interface{})
	if len(messages) != 1 {
		t.Errorf("expected 1 message (system excluded), got %d", len(messages))
	}
}

func TestClaudeProvider_Chat_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid api key"}`))
	}))
	defer server.Close()

	p := claude.NewProvider("bad-key", "model", claude.WithBaseURL(server.URL))
	_, err := p.Chat(context.Background(), []ai.Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Error("expected error for 401 response, got nil")
	}
}

func TestClaudeProvider_Name(t *testing.T) {
	p := claude.NewProvider("key", "model")
	if p.Name() != "claude" {
		t.Errorf("expected claude, got %s", p.Name())
	}
}

func TestClaudeProvider_Stream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": "流式回复"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := claude.NewProvider("key", "model", claude.WithBaseURL(server.URL))
	ch, err := p.Stream(context.Background(), []ai.Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatal(err)
	}

	var result string
	for chunk := range ch {
		if chunk.Err != nil {
			t.Fatal(chunk.Err)
		}
		result += chunk.Content
	}
	if result != "流式回复" {
		t.Errorf("expected '流式回复', got %s", result)
	}
}

// 编译时验证 *Provider 实现了 ai.Provider 接口
var _ ai.Provider = (*claude.Provider)(nil)
