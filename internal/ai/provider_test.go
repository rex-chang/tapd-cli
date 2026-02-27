package ai_test

import (
	"context"
	"testing"

	"github.com/rex-chang/tapd-cli/internal/ai"
)

// mockProvider 用于测试
type mockProvider struct {
	response string
}

func (m *mockProvider) Name() string { return "mock" }
func (m *mockProvider) Chat(ctx context.Context, messages []ai.Message) (string, error) {
	return m.response, nil
}
func (m *mockProvider) Stream(ctx context.Context, messages []ai.Message) (<-chan ai.StreamChunk, error) {
	ch := make(chan ai.StreamChunk, 1)
	ch <- ai.StreamChunk{Content: m.response, Done: true}
	close(ch)
	return ch, nil
}

// 编译时验证 mockProvider 实现了 Provider 接口
var _ ai.Provider = (*mockProvider)(nil)

func TestProvider_Interface(t *testing.T) {
	p := &mockProvider{response: "hello"}

	resp, err := p.Chat(context.Background(), []ai.Message{
		{Role: "user", Content: "test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp != "hello" {
		t.Errorf("expected hello, got %s", resp)
	}
}

func TestProvider_Stream(t *testing.T) {
	p := &mockProvider{response: "streamed"}

	ch, err := p.Stream(context.Background(), []ai.Message{
		{Role: "user", Content: "test"},
	})
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

	if result != "streamed" {
		t.Errorf("expected streamed, got %s", result)
	}
}
