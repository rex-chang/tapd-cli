# tapd-cli AI 友好化实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将 tapd-cli 改造为 AI 友好项目，新增上下分屏 TUI 布局和可插拔多 Provider AI 对话功能。

**Architecture:** 新增 `internal/ai/` 模块提供抽象 Provider 接口，TUI 层扩展为上下分屏（数据区 + 对话区），通过 Context 感知机制将当前选中的 TAPD 条目自动注入 AI 系统 Prompt。

**Tech Stack:** Go 1.25+, Bubble Tea v1.3.10, Bubbles v1.0.0 (textarea + viewport), Lipgloss v1.1.0, net/http (Anthropic API)

---

## 背景：现有代码结构

```
internal/config/config.go   # Config struct + LoadConfig()
internal/tapd/client.go     # tapd.Client + DoGet()
internal/tapd/workspace.go  # GetWorkspaces()
internal/tui/app.go         # Bubble Tea Model（单区域布局）
cmd/tapd-cli/main.go        # 入口
```

现有 `Config` 结构体：
```go
type Config struct {
    APIUser     string `yaml:"api_user"`
    APIPassword string `yaml:"api_password"`
    APIToken    string `yaml:"api_token"`
    WorkspaceID string `yaml:"workspace_id"`
}
```

---

## Task 1: 扩展 Config 以支持 AI 配置

**Files:**
- Modify: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Step 1: 写失败测试**

在 `internal/config/config_test.go` 中写：

```go
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

    // AI 配置可选，未配置时不报错
    if cfg.AI.Provider != "" {
        t.Errorf("expected empty provider, got %s", cfg.AI.Provider)
    }
}
```

**Step 2: 运行测试确认失败**

```bash
go test ./internal/config/... -v -run TestLoadConfig
```

期望：FAIL，`config.LoadConfigFromPath undefined`

**Step 3: 实现最小代码**

在 `internal/config/config.go` 中添加：

```go
// AIConfig 存储 AI Provider 相关配置
type AIConfig struct {
    Provider string `yaml:"provider"` // claude | openai
    APIKey   string `yaml:"api_key"`
    Model    string `yaml:"model"`
}

// 在 Config struct 中添加字段
type Config struct {
    APIUser     string    `yaml:"api_user"`
    APIPassword string    `yaml:"api_password"`
    APIToken    string    `yaml:"api_token"`
    WorkspaceID string    `yaml:"workspace_id"`
    AI          AIConfig  `yaml:"ai"`
}

// LoadConfigFromPath 从指定路径加载配置（方便测试）
func LoadConfigFromPath(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("读取配置文件失败: %w", err)
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("解析配置文件失败: %w", err)
    }

    if cfg.APIUser == "" || cfg.WorkspaceID == "" {
        return nil, fmt.Errorf("配置文件中必需包含 api_user 和 workspace_id")
    }
    if cfg.APIPassword == "" && cfg.APIToken == "" {
        return nil, fmt.Errorf("配置文件中必需提供 api_password 或 api_token 之一进行身份验证")
    }

    return &cfg, nil
}
```

同时重构 `LoadConfig()` 复用 `LoadConfigFromPath()`：

```go
func LoadConfig() (*Config, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return nil, fmt.Errorf("无法获取用户主目录: %w", err)
    }

    configDir := filepath.Join(homeDir, ".tapd-cli")
    configPath := filepath.Join(configDir, "config.yaml")

    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        if err := promptConfig(configDir, configPath); err != nil {
            return nil, fmt.Errorf("初始化配置失败: %w", err)
        }
    }

    return LoadConfigFromPath(configPath)
}
```

**Step 4: 运行测试确认通过**

```bash
go test ./internal/config/... -v -run TestLoadConfig
```

期望：PASS

**Step 5: 确认整体编译通过**

```bash
go build ./...
```

**Step 6: 提交**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat: Config 扩展 AIConfig，新增 LoadConfigFromPath 方法"
```

---

## Task 2: AI Provider 抽象接口

**Files:**
- Create: `internal/ai/provider.go`
- Create: `internal/ai/provider_test.go`

**Step 1: 写失败测试（验证 mock 实现接口）**

创建 `internal/ai/provider_test.go`：

```go
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
```

**Step 2: 运行测试确认失败**

```bash
go test ./internal/ai/... -v
```

期望：FAIL，`package ai not found`

**Step 3: 创建 provider.go**

创建 `internal/ai/provider.go`：

```go
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
```

**Step 4: 运行测试确认通过**

```bash
go test ./internal/ai/... -v
```

期望：PASS

**Step 5: 提交**

```bash
git add internal/ai/provider.go internal/ai/provider_test.go
git commit -m "feat: 新增 AI Provider 抽象接口"
```

---

## Task 3: Claude Provider 实现

**Files:**
- Create: `internal/ai/claude/claude.go`
- Create: `internal/ai/claude/claude_test.go`

**Step 1: 写失败测试（使用 httptest mock API）**

创建 `internal/ai/claude/claude_test.go`：

```go
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
    // Mock Anthropic API
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Header.Get("x-api-key") == "" {
            t.Error("missing x-api-key header")
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

func TestClaudeProvider_Name(t *testing.T) {
    p := claude.NewProvider("key", "model")
    if p.Name() != "claude" {
        t.Errorf("expected claude, got %s", p.Name())
    }
}
```

**Step 2: 运行测试确认失败**

```bash
go test ./internal/ai/claude/... -v
```

期望：FAIL，`package claude not found`

**Step 3: 创建 claude.go**

创建 `internal/ai/claude/claude.go`：

```go
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
        return "", err
    }

    req, err := http.NewRequestWithContext(ctx, "POST",
        p.baseURL+"/v1/messages", bytes.NewReader(data))
    if err != nil {
        return "", err
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

// Stream 流式调用（当前实现为非流式的兼容包装，后续可扩展为 SSE）
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
```

**Step 4: 运行测试确认通过**

```bash
go test ./internal/ai/claude/... -v
```

期望：PASS

**Step 5: 确认编译**

```bash
go build ./...
```

**Step 6: 提交**

```bash
git add internal/ai/claude/claude.go internal/ai/claude/claude_test.go
git commit -m "feat: 新增 Claude AI Provider 实现"
```

---

## Task 4: TUI 分屏布局与 Chat 组件

**Files:**
- Create: `internal/tui/chat.go`
- Modify: `internal/tui/app.go`

**Step 1: 写失败测试（验证 Chat 组件基础行为）**

创建 `internal/tui/chat_test.go`：

```go
package tui_test

import (
    "testing"

    "github.com/rex-chang/tapd-cli/internal/tui"
)

func TestChatModel_AddMessage(t *testing.T) {
    m := tui.NewChatModel(80, 20)

    m.AddMessage("user", "你好")
    m.AddMessage("assistant", "你好！有什么可以帮助你的？")

    history := m.History()
    if len(history) != 2 {
        t.Errorf("expected 2 messages, got %d", len(history))
    }
    if history[0].Role != "user" {
        t.Errorf("expected user, got %s", history[0].Role)
    }
}

func TestChatModel_Clear(t *testing.T) {
    m := tui.NewChatModel(80, 20)
    m.AddMessage("user", "test")
    m.Clear()

    if len(m.History()) != 0 {
        t.Error("expected empty history after clear")
    }
}
```

**Step 2: 运行测试确认失败**

```bash
go test ./internal/tui/... -v -run TestChatModel
```

期望：FAIL，`tui.NewChatModel undefined`

**Step 3: 创建 chat.go**

创建 `internal/tui/chat.go`：

```go
package tui

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/bubbles/textarea"
    "github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"

    "github.com/rex-chang/tapd-cli/internal/ai"
)

var (
    // 对话区样式
    chatBorderStyle = lipgloss.NewStyle().
        Border(lipgloss.NormalBorder(), true, false, false, false).
        BorderForeground(lipgloss.Color("240"))

    userMsgStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("86")).
        Bold(true)

    assistantMsgStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("252"))

    inputStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("62")).
        Padding(0, 1)
)

// ChatMessage 存储一条对话记录
type ChatMessage struct {
    Role    string
    Content string
}

// ChatModel 是底部 AI 对话区的 Bubble Tea 组件
type ChatModel struct {
    viewport viewport.Model
    input    textarea.Model
    history  []ChatMessage
    width    int
    height   int
    focused  bool
    thinking bool // AI 是否正在响应
}

// aiResponseMsg AI 响应完成的消息
type aiResponseMsg struct {
    content string
    err     error
}

// NewChatModel 创建聊天组件
func NewChatModel(width, height int) ChatModel {
    ta := textarea.New()
    ta.Placeholder = "输入消息... (Enter 发送, Shift+Enter 换行)"
    ta.Focus()
    ta.SetWidth(width - 4)
    ta.SetHeight(3)
    ta.ShowLineNumbers = false
    ta.CharLimit = 2000

    vp := viewport.New(width, height-6) // 扣除输入框高度
    vp.SetContent("")

    return ChatModel{
        viewport: vp,
        input:    ta,
        width:    width,
        height:   height,
    }
}

func (m ChatModel) Init() tea.Cmd {
    return textarea.Blink
}

func (m ChatModel) Update(msg tea.Msg) (ChatModel, tea.Cmd) {
    var (
        vpCmd  tea.Cmd
        taCmd  tea.Cmd
    )

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.Type {
        case tea.KeyEnter:
            if !msg.Alt { // 普通 Enter 发送，Alt+Enter 换行
                userInput := strings.TrimSpace(m.input.Value())
                if userInput == "" {
                    return m, nil
                }
                m.AddMessage("user", userInput)
                m.input.Reset()
                m.thinking = true
                m.refreshViewport()
                // 发送消息的 Cmd 由父组件处理（需要访问 AI Provider）
                return m, sendMessageCmd(userInput)
            }
        }

    case aiResponseMsg:
        m.thinking = false
        if msg.err != nil {
            m.AddMessage("assistant", fmt.Sprintf("错误: %v", msg.err))
        } else {
            m.AddMessage("assistant", msg.content)
        }
        m.refreshViewport()
    }

    m.viewport, vpCmd = m.viewport.Update(msg)
    m.input, taCmd = m.input.Update(msg)

    return m, tea.Batch(vpCmd, taCmd)
}

func (m ChatModel) View() string {
    header := chatBorderStyle.Width(m.width).Render(" AI 助手 (Tab 切换焦点, q 退出)")

    inputView := inputStyle.Width(m.width - 2).Render(m.input.View())

    return lipgloss.JoinVertical(lipgloss.Left,
        header,
        m.viewport.View(),
        inputView,
    )
}

// AddMessage 添加一条消息到历史记录
func (m *ChatModel) AddMessage(role, content string) {
    m.history = append(m.history, ChatMessage{Role: role, Content: content})
    m.refreshViewport()
}

// History 返回对话历史
func (m *ChatModel) History() []ChatMessage {
    return m.history
}

// Clear 清空对话历史
func (m *ChatModel) Clear() {
    m.history = nil
    m.viewport.SetContent("")
}

// SetSize 更新组件尺寸
func (m *ChatModel) SetSize(width, height int) {
    m.width = width
    m.height = height
    m.viewport.Width = width
    m.viewport.Height = height - 6
    m.input.SetWidth(width - 4)
    m.refreshViewport()
}

// Focus 聚焦输入框
func (m *ChatModel) Focus() {
    m.focused = true
    m.input.Focus()
}

// Blur 取消聚焦
func (m *ChatModel) Blur() {
    m.focused = false
    m.input.Blur()
}

// refreshViewport 刷新消息列表显示
func (m *ChatModel) refreshViewport() {
    var sb strings.Builder
    for _, msg := range m.history {
        switch msg.Role {
        case "user":
            sb.WriteString(userMsgStyle.Render("你: "))
            sb.WriteString(msg.Content)
        case "assistant":
            sb.WriteString(assistantMsgStyle.Render("AI: "))
            sb.WriteString(msg.Content)
        }
        sb.WriteString("\n\n")
    }

    if m.thinking {
        sb.WriteString(assistantMsgStyle.Render("AI: ") + "正在思考...")
    }

    m.viewport.SetContent(sb.String())
    m.viewport.GotoBottom()
}

// sendMessageCmd 是一个占位 Cmd，父组件监听后执行实际 AI 调用
type sendMessageCmd string

func (s sendMessageCmd) isCmd() {} // 标识符，无实际作用
```

> 注意：`sendMessageCmd` 是自定义类型，父组件（app.go）的 Update 方法中需要 type switch 拦截它。

**Step 4: 运行测试确认通过**

```bash
go test ./internal/tui/... -v -run TestChatModel
```

期望：PASS

**Step 5: 修改 app.go 集成分屏布局**

修改 `internal/tui/app.go`，将 `Model` 扩展为分屏布局：

**在 import 中新增：**
```go
"github.com/rex-chang/tapd-cli/internal/ai"
"github.com/rex-chang/tapd-cli/internal/config"  // 已有
```

**在 Model struct 中新增字段：**
```go
type Model struct {
    list       list.Model
    spinner    spinner.Model
    chat       ChatModel   // 新增：底部对话区
    provider   ai.Provider // 新增：AI Provider（可为 nil）
    config     *config.Config
    client     *tapd.Client
    state      viewState
    focus      focusArea   // 新增：当前焦点区域
    quitting   bool
    err        error
    termWidth  int
    termHeight int
}

// focusArea 标识当前焦点在哪个区域
type focusArea int

const (
    focusData focusArea = iota // 焦点在上方数据区
    focusChat                  // 焦点在下方对话区
)
```

**修改 `InitialModel` 函数签名，接受可选 Provider：**
```go
func InitialModel(cfg *config.Config, provider ai.Provider) Model {
    // ... 原有初始化代码 ...
    chat := NewChatModel(80, 15)

    return Model{
        list:     m,
        spinner:  s,
        chat:     chat,
        provider: provider,
        config:   cfg,
        client:   client,
        state:    stateLoading,
        focus:    focusData,
    }
}
```

**在 `Update` 中添加 Tab 切换焦点和 AI 调用逻辑：**

在 `case tea.KeyMsg:` 分支前，添加 `sendMessageCmd` 处理：
```go
case sendMessageCmd:
    if m.provider != nil {
        userMsg := string(msg)
        // 构建带上下文的消息列表
        messages := m.buildContextMessages(userMsg)
        return m, callAICmd(m.provider, messages)
    }
    m.chat.AddMessage("assistant", "未配置 AI Provider，请在 ~/.tapd-cli/config.yaml 中添加 ai 配置")
    return m, nil

case aiCallResultMsg:
    m.chat, _ = m.chat.Update(aiResponseMsg{content: msg.content, err: msg.err})
    return m, nil
```

在 `case tea.KeyMsg:` 中添加 Tab 处理：
```go
case tea.KeyTab:
    if m.focus == focusData {
        m.focus = focusChat
        m.list.SetFilteringEnabled(false)
        m.chat.Focus()
    } else {
        m.focus = focusData
        m.chat.Blur()
    }
    return m, nil
```

在 `case tea.WindowSizeMsg:` 中同步更新 chat 尺寸：
```go
case tea.WindowSizeMsg:
    m.termWidth = msg.Width
    m.termHeight = msg.Height
    chatHeight := msg.Height * 30 / 100 // 下方占 30%
    dataHeight := msg.Height - chatHeight
    h, v := docStyle.GetFrameSize()
    m.list.SetSize(msg.Width-h, dataHeight-v-3)
    m.chat.SetSize(msg.Width-h, chatHeight)
```

**修改 `View()` 支持分屏渲染：**
```go
func (m Model) View() string {
    if m.quitting {
        return "再见！\n"
    }

    header := infoStyle.Render(fmt.Sprintf("用户: %s | Workspace: %s | Tab 切换焦点",
        m.config.APIUser, m.config.WorkspaceID))

    var dataContent string
    if m.err != nil {
        dataContent = errStyle.Render(fmt.Sprintf("加载失败: %v", m.err))
    } else if m.state == stateLoading {
        dataContent = fmt.Sprintf("\n\n   %s 正在加载 TAPD 项目...", m.spinner.View())
    } else {
        dataContent = m.list.View()
    }

    dataPane := header + "\n\n" + dataContent
    chatPane := m.chat.View()

    return docStyle.Render(
        lipgloss.JoinVertical(lipgloss.Left, dataPane, chatPane),
    )
}
```

**新增辅助函数（在 app.go 底部）：**
```go
// buildContextMessages 构建含上下文的消息列表
func (m Model) buildContextMessages(userMsg string) []ai.Message {
    var systemPrompt strings.Builder
    systemPrompt.WriteString("你是 TAPD 项目管理助手，帮助用户分析需求和缺陷。\n\n")

    // 自动注入当前选中的 TAPD 条目上下文
    if m.state == stateList {
        if selected := m.list.SelectedItem(); selected != nil {
            if it, ok := selected.(item); ok {
                systemPrompt.WriteString(fmt.Sprintf("当前用户正在查看的条目：\n标题: %s\n描述: %s\n",
                    it.title, it.desc))
            }
        }
    }

    messages := []ai.Message{
        {Role: "system", Content: systemPrompt.String()},
    }
    // 附加历史对话
    for _, h := range m.chat.History() {
        messages = append(messages, ai.Message{Role: h.Role, Content: h.Content})
    }
    // 当前用户输入
    messages = append(messages, ai.Message{Role: "user", Content: userMsg})
    return messages
}

// aiCallResultMsg AI 异步调用结果
type aiCallResultMsg struct {
    content string
    err     error
}

// callAICmd 异步调用 AI Provider
func callAICmd(p ai.Provider, messages []ai.Message) tea.Cmd {
    return func() tea.Msg {
        resp, err := p.Chat(context.Background(), messages)
        return aiCallResultMsg{content: resp, err: err}
    }
}
```

**Step 6: 修改 main.go 传入 Provider**

```go
func main() {
    cfg, err := config.LoadConfig()
    if err != nil {
        fmt.Printf("配置错误: %v\n", err)
        os.Exit(1)
    }

    // 根据配置初始化 AI Provider（可选）
    var provider ai.Provider
    if cfg.AI.APIKey != "" {
        switch cfg.AI.Provider {
        case "claude", "":
            model := cfg.AI.Model
            if model == "" {
                model = "claude-3-5-sonnet-20241022"
            }
            provider = claude.NewProvider(cfg.AI.APIKey, model)
        }
    }

    model := tui.InitialModel(cfg, provider)
    p := tea.NewProgram(model, tea.WithAltScreen())

    if _, err := p.Run(); err != nil {
        fmt.Printf("启动 tapd-cli 失败: %v\n", err)
        os.Exit(1)
    }
}
```

**main.go 需要新增 import：**
```go
import (
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"

    "github.com/rex-chang/tapd-cli/internal/ai"
    "github.com/rex-chang/tapd-cli/internal/ai/claude"
    "github.com/rex-chang/tapd-cli/internal/config"
    "github.com/rex-chang/tapd-cli/internal/tui"
)
```

**Step 7: 编译运行验证**

```bash
go build -o bin/tapd-cli ./cmd/tapd-cli
./bin/tapd-cli
```

期望：正常启动，Tab 键可切换焦点到下方对话区，输入框可用。

**Step 8: 提交**

```bash
git add internal/tui/chat.go internal/tui/chat_test.go internal/tui/app.go cmd/tapd-cli/main.go
git commit -m "feat: 分屏 TUI 布局，新增底部 AI 对话区组件"
```

---

## Task 5: 快捷键完善与 AI 功能指令

**Files:**
- Modify: `internal/tui/chat.go`

**Step 1: 在 chat.go 中添加 `/summarize`、`/search`、`/draft` 指令处理**

在 `chat.go` 的 `Update` 函数的 `tea.KeyEnter` 处理中，在 `sendMessageCmd` 前解析指令：

```go
case tea.KeyEnter:
    if !msg.Alt {
        userInput := strings.TrimSpace(m.input.Value())
        if userInput == "" {
            return m, nil
        }
        m.AddMessage("user", userInput)
        m.input.Reset()
        m.thinking = true
        m.refreshViewport()

        // 解析内置指令，转换为对 AI 的自然语言请求
        expanded := expandCommand(userInput)
        return m, sendMessageCmd(expanded)
    }
```

在文件底部新增 `expandCommand` 函数：

```go
// expandCommand 将 / 指令扩展为对 AI 的完整指令
func expandCommand(input string) string {
    switch {
    case input == "/summarize" || input == "/s":
        return "请总结当前查看的条目，提炼核心要点、目标和影响范围，用简洁的中文列表呈现。"
    case strings.HasPrefix(input, "/search "):
        query := strings.TrimPrefix(input, "/search ")
        return fmt.Sprintf("请将以下需求转换为 TAPD 搜索条件：%s。输出搜索关键词和筛选条件。", query)
    case input == "/draft" || input == "/d":
        return "请根据当前条目的信息，帮我起草一份专业的需求描述或缺陷报告模板。"
    case input == "/help":
        return "请列出你能帮我做的事情，包括总结条目、搜索、写作等功能。"
    }
    return input
}
```

**Step 2: 运行测试**

```bash
go test ./internal/tui/... -v
go build ./...
```

**Step 3: 提交**

```bash
git add internal/tui/chat.go
git commit -m "feat: 支持 /summarize /search /draft 快捷指令"
```

---

## Task 6: 完善 AI 开发基础设施

**Files:**
- Create: `docs/architecture.md`
- Create: `.agent/workflows/feature.md`
- Modify: `CLAUDE.md`

**Step 1: 创建 architecture.md**

创建 `docs/architecture.md`，内容如下（参考下方 Step 3 的 CLAUDE.md 更新部分）：

记录所有模块的职责、接口和依赖关系，以便 AI 助手快速理解代码结构。

```markdown
# tapd-cli 架构文档

## 模块依赖图

main.go
  → config.LoadConfig()        # 加载配置，含可选 AI 配置
  → ai.Provider (可选)         # 根据配置初始化 Claude/OpenAI Provider
  → tui.InitialModel(cfg, p)   # 注入配置和 Provider，创建 TUI 模型
  → tea.NewProgram(model)      # 启动 Bubble Tea 事件循环

## 各模块职责

### internal/config
- 单一职责：从 ~/.tapd-cli/config.yaml 加载配置
- 对外接口：LoadConfig(), LoadConfigFromPath(path)
- 测试：使用 t.TempDir() 创建临时配置文件

### internal/tapd
- 单一职责：与 TAPD REST API 通信
- 对外接口：NewClient(cfg), Client.DoGet(path, query), Client.GetWorkspaces()
- 扩展规范：新 API 端点新建单独文件（如 story.go, bug.go）

### internal/ai
- 单一职责：定义 AI Provider 抽象接口
- 对外接口：Provider interface, Message struct, StreamChunk struct
- 扩展规范：新 Provider 在 internal/ai/<name>/ 子包实现

### internal/ai/claude
- 单一职责：Anthropic Claude API 实现
- 测试：使用 httptest.NewServer mock API

### internal/tui
- 单一职责：Bubble Tea 分屏 TUI 应用
- 核心模型：Model（上方数据区）+ ChatModel（下方对话区）
- 焦点切换：Tab 键在 focusData / focusChat 间切换
- AI 调用流：sendMessageCmd → parent Update → callAICmd → aiCallResultMsg

## 消息流图（TUI）

用户按 Enter
  → chat.Update() 产生 sendMessageCmd
  → app.Update() 拦截，调用 callAICmd(provider, messages)
  → AI Provider.Chat() 异步执行
  → 返回 aiCallResultMsg
  → app.Update() 转发给 chat.Update(aiResponseMsg)
  → chat.refreshViewport() 更新显示
```

**Step 2: 创建 `.agent/workflows/feature.md`**

```bash
mkdir -p /Users/rexchang/learn/tapd-cli/.agent/workflows
```

创建 `.agent/workflows/feature.md`：

```markdown
# 新功能开发工作流

## 标准流程

1. 阅读 docs/context.md 和 docs/architecture.md 了解项目背景
2. 在 docs/plans/ 创建设计文档（格式：YYYY-MM-DD-<feature>.md）
3. 遵循 TDD：先写测试，运行确认失败，再实现，运行确认通过
4. 小步提交：每个独立功能单元提交一次
5. 更新 docs/architecture.md（如有架构变化）

## 代码规范

- 所有注释使用中文
- 函数/类型名使用英文，符合 Go 惯例
- 错误信息使用中文，方便用户阅读
- 接口优先设计：先定义接口，再实现
- 测试文件使用 _test.go 后缀，包名加 _test

## 目录规范

- 新 TAPD API 端点：internal/tapd/<resource>.go
- 新 AI Provider：internal/ai/<provider>/
- 新 TUI 视图/组件：internal/tui/<component>.go
- 设计文档：docs/plans/YYYY-MM-DD-<topic>.md
```

**Step 3: 更新 CLAUDE.md**

在 CLAUDE.md 的"常用命令"部分追加：

```markdown
# 运行单个测试
go test ./internal/config/... -v -run TestLoadConfig_WithAI

# 运行所有测试
go test ./... -v
```

在"技术栈"后新增"AI 集成"部分，记录 Provider 接口和扩展方式。

**Step 4: 提交**

```bash
git add docs/architecture.md .agent/workflows/feature.md CLAUDE.md
git commit -m "docs: 新增架构文档和 Agent 工作流，完善 AI 开发基础设施"
```

---

## 验收标准

全部完成后执行以下验证：

```bash
# 1. 所有测试通过
go test ./... -v

# 2. 编译成功
go build -o bin/tapd-cli ./cmd/tapd-cli

# 3. 手动测试
./bin/tapd-cli
# 期望：
# - 程序正常启动，显示 TAPD 项目列表
# - Tab 键切换焦点到下方 AI 对话框
# - 输入消息后按 Enter，若已配置 AI Provider，返回 AI 回复
# - 未配置 Provider 时，显示友好提示
# - /summarize 指令触发 AI 总结当前条目
```

---

## 配置示例（~/.tapd-cli/config.yaml）

```yaml
api_user: your_tapd_user
api_token: your_tapd_token
workspace_id: "12345678"

ai:
  provider: claude
  api_key: sk-ant-xxxxx
  model: claude-3-5-sonnet-20241022
```
