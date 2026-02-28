package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// 对话区顶部分隔线样式
	chatBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(lipgloss.Color("240"))

	// 用户消息标签样式
	userMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

	// AI 消息样式
	assistantMsgStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	// 输入框边框样式
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

// aiResponseMsg AI 响应完成的消息（由父组件 app.go 发送给 ChatModel）
type aiResponseMsg struct {
	content string
	err     error
}

// sendMessageCmd 是一个自定义 Cmd 类型，
// 父组件（app.go）的 Update 方法拦截后执行实际 AI 调用
type sendMessageCmd string

// ChatModel 是底部 AI 对话区的 Bubble Tea 组件
type ChatModel struct {
	viewport viewport.Model
	input    textarea.Model
	history  []ChatMessage
	width    int
	height   int
	focused  bool
	thinking bool // AI 是否正在响应中
}

// NewChatModel 创建聊天组件，width/height 为初始尺寸
func NewChatModel(width, height int) ChatModel {
	ta := textarea.New()
	ta.Placeholder = "输入消息... (Enter 发送, Alt+Enter 换行)"
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

// Init 初始化 ChatModel（启动光标闪烁）
func (m ChatModel) Init() tea.Cmd {
	return textarea.Blink
}

// Update 处理 ChatModel 的消息
func (m ChatModel) Update(msg tea.Msg) (ChatModel, tea.Cmd) {
	var (
		vpCmd tea.Cmd
		taCmd tea.Cmd
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
				// 展开内置指令，发送给父组件处理
				expanded := expandCommand(userInput)
				return m, func() tea.Msg { return sendMessageCmd(expanded) }
			}
		}

	case aiResponseMsg:
		// 接收 AI 响应
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

// View 渲染聊天区域
func (m ChatModel) View() string {
	header := chatBorderStyle.Width(m.width).Render(" AI 助手 (Tab 切换焦点)")

	inputView := inputStyle.Width(m.width - 2).Render(m.input.View())

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		m.viewport.View(),
		inputView,
	)
}

// AddMessage 添加一条消息到历史记录（同时刷新 viewport）
func (m *ChatModel) AddMessage(role, content string) {
	m.history = append(m.history, ChatMessage{Role: role, Content: content})
	m.refreshViewport()
}

// History 返回对话历史（只读）
func (m *ChatModel) History() []ChatMessage {
	return m.history
}

// Clear 清空对话历史和显示内容
func (m *ChatModel) Clear() {
	m.history = nil
	m.viewport.SetContent("")
}

// SetSize 更新组件尺寸（终端 resize 时调用）
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

// refreshViewport 重新构建消息列表显示内容
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

// expandCommand 将 /指令 扩展为对 AI 的完整自然语言请求
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
