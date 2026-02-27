package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rex-chang/tapd-cli/internal/ai"
)

const defaultBaseURL = "https://api.anthropic.com"

// Provider 是 Anthropic Claude 的 AI Provider 实现
type Provider struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

type option func(*Provider)

// WithBaseURL 覆盖 API Base URL（用于测试）
func WithBaseURL(url string) option {
	return func(p *Provider) { p.baseURL = url }
}

// NewProvider 创建 Claude Provider
func NewProvider(apiKey, model string, opts ...option) *Provider {
	p := &Provider{
		apiKey:  apiKey,
		model:   model,
		baseURL: defaultBaseURL,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *Provider) Name() string { return "claude" }

// Chat 调用 Anthropic Messages API 获取完整响应
func (p *Provider) Chat(ctx context.Context, messages []ai.Message) (string, error) {
	// 分离 system prompt 和对话消息
	var systemPrompt string
	var chatMessages []map[string]string

	for _, m := range messages {
		if m.Role == "system" {
			systemPrompt = m.Content
			continue
		}
		chatMessages = append(chatMessages, map[string]string{
			"role":    m.Role,
			"content": m.Content,
		})
	}

	body := map[string]interface{}{
		"model":      p.model,
		"max_tokens": 2048,
		"messages":   chatMessages,
	}
	if systemPrompt != "" {
		body["system"] = systemPrompt
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		p.baseURL+"/v1/messages", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求 Claude API 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Claude API 错误 (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	for _, block := range result.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}
	return "", fmt.Errorf("响应中未找到文本内容")
}

// Stream 流式调用（当前为非流式兼容包装，后续可扩展为 SSE）
func (p *Provider) Stream(ctx context.Context, messages []ai.Message) (<-chan ai.StreamChunk, error) {
	ch := make(chan ai.StreamChunk, 1)
	go func() {
		defer close(ch)
		resp, err := p.Chat(ctx, messages)
		if err != nil {
			ch <- ai.StreamChunk{Err: err, Done: true}
			return
		}
		ch <- ai.StreamChunk{Content: resp, Done: true}
	}()
	return ch, nil
}
