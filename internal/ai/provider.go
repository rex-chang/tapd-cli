package ai

import "context"

// Message 表示对话中的一条消息
type Message struct {
	Role    string // "system" | "user" | "assistant"
	Content string
}

// StreamChunk 流式响应的单个分片
type StreamChunk struct {
	Content string
	Done    bool
	Err     error
}

// Provider 是所有 AI 服务的抽象接口
type Provider interface {
	// Name 返回 Provider 名称（如 "claude", "openai"）
	Name() string
	// Chat 发送消息列表，返回完整响应
	Chat(ctx context.Context, messages []Message) (string, error)
	// Stream 发送消息列表，流式返回响应分片
	Stream(ctx context.Context, messages []Message) (<-chan StreamChunk, error)
}
