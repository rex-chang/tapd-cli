package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/rex-chang/tapd-cli/internal/ai"
	"github.com/rex-chang/tapd-cli/internal/config"
	"github.com/rex-chang/tapd-cli/internal/tapd"
)

var (
	docStyle  = lipgloss.NewStyle().Margin(1, 2)
	infoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginBottom(1)
	errStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true).MarginBottom(1)
)

type item struct {
	title, desc, fullDesc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

// viewState 表示数据区的当前状态
type viewState int

const (
	stateLoading viewState = iota // 数据加载中
	stateList                     // 显示项目列表
)

// focusArea 标识当前焦点在哪个区域
type focusArea int

const (
	focusData focusArea = iota // 焦点在上方数据区
	focusChat                  // 焦点在下方对话区
)

// storiesLoadedMsg TAPD 需求加载完成消息
type storiesLoadedMsg []tapd.StoryItem

// errMsg 错误消息
type errMsg struct{ err error }

// aiCallResultMsg AI 异步调用结果消息
type aiCallResultMsg struct {
	content string
	err     error
}

// Model 是整个 TUI 应用的根模型，包含上方数据区和下方对话区
type Model struct {
	list       list.Model
	spinner    spinner.Model
	chat       ChatModel
	provider   ai.Provider // 可为 nil（未配置时降级提示）
	config     *config.Config
	client     *tapd.Client
	state      viewState
	focus      focusArea
	quitting   bool
	err        error
	termWidth  int
	termHeight int
}

// InitialModel 创建初始 TUI 模型，provider 可为 nil
func InitialModel(cfg *config.Config, provider ai.Provider) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "TAPD 需求列表"
	l.SetShowTitle(true)

	client := tapd.NewClient(cfg)
	chat := NewChatModel(80, 15)

	return Model{
		list:     l,
		spinner:  s,
		chat:     chat,
		provider: provider,
		config:   cfg,
		client:   client,
		state:    stateLoading,
		focus:    focusData,
	}
}

// fetchStories 在后台异步拉取 TAPD 需求列表
func fetchStories(client *tapd.Client) tea.Cmd {
	return func() tea.Msg {
		stories, err := client.GetStories()
		if err != nil {
			return errMsg{err: err}
		}
		return storiesLoadedMsg(stories)
	}
}

// callAICmd 异步调用 AI Provider，结果通过 aiCallResultMsg 返回
func callAICmd(p ai.Provider, messages []ai.Message) tea.Cmd {
	return func() tea.Msg {
		resp, err := p.Chat(context.Background(), messages)
		return aiCallResultMsg{content: resp, err: err}
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, fetchStories(m.client))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case errMsg:
		m.err = msg.err
		m.state = stateList
		return m, nil

	case storiesLoadedMsg:
		var items []list.Item
		for _, s := range msg {
			items = append(items, item{
				title:    s.Story.Name,
				desc:     fmt.Sprintf("[%s] 创建人: %s | ID: %s", s.Story.Status, s.Story.Creator, s.Story.ID),
				fullDesc: s.Story.Description,
			})
		}
		cmd := m.list.SetItems(items)
		m.state = stateList
		return m, cmd

	// 拦截 chat 发出的 sendMessageCmd，执行实际 AI 调用
	case sendMessageCmd:
		if m.provider != nil {
			messages := m.buildContextMessages(string(msg))
			return m, callAICmd(m.provider, messages)
		}
		// 未配置 Provider 时给出友好提示
		m.chat.AddMessage("assistant", "未配置 AI Provider，请在 ~/.tapd-cli/config.yaml 中添加 ai 配置")
		return m, nil

	// AI 调用结果，转发给 ChatModel 处理
	case aiCallResultMsg:
		var cmd tea.Cmd
		m.chat, cmd = m.chat.Update(aiResponseMsg{content: msg.content, err: msg.err})
		return m, cmd

	case tea.KeyMsg:
		// q / Ctrl+C 退出
		if msg.String() == "ctrl+c" || (msg.String() == "q" && m.focus == focusData) {
			m.quitting = true
			return m, tea.Quit
		}
		// Tab 切换焦点
		if msg.Type == tea.KeyTab {
			if m.focus == focusData {
				m.focus = focusChat
				m.list.SetFilteringEnabled(false)
				m.chat.Focus()
			} else {
				m.focus = focusData
				m.chat.Blur()
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		chatHeight := msg.Height * 30 / 100 // 对话区占 30%
		dataHeight := msg.Height - chatHeight
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, dataHeight-v-3)
		m.chat.SetSize(msg.Width-h*2, chatHeight)
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	// 根据焦点决定将消息路由给哪个组件
	if m.focus == focusChat {
		m.chat, cmd = m.chat.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		if m.state == stateLoading {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		} else if m.state == stateList {
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.quitting {
		return "再见！\n"
	}

	header := infoStyle.Render(fmt.Sprintf("Workspace: %s | Tab 切换焦点",
		m.config.WorkspaceID))

	var dataContent string
	if m.err != nil {
		dataContent = errStyle.Render(fmt.Sprintf("加载失败: %v", m.err))
	} else if m.state == stateLoading {
		dataContent = fmt.Sprintf("\n\n   %s 正在加载 TAPD 需求...", m.spinner.View())
	} else {
		dataContent = m.list.View()
	}

	dataPane := header + "\n\n" + dataContent
	chatPane := m.chat.View()

	return docStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, dataPane, chatPane),
	)
}

// buildContextMessages 构建包含当前 TAPD 上下文的消息列表
func (m Model) buildContextMessages(userMsg string) []ai.Message {
	var systemPrompt strings.Builder
	systemPrompt.WriteString("你是 TAPD 项目管理助手，帮助用户分析需求和缺陷。\n\n")

	// 自动注入当前选中条目作为上下文
	if m.state == stateList {
		if selected := m.list.SelectedItem(); selected != nil {
			if it, ok := selected.(item); ok {
				systemPrompt.WriteString(fmt.Sprintf(
					"当前用户正在查看的条目：\n标题: %s\n状态属性: %s\n详情描述: %s\n",
					it.title, it.desc, it.fullDesc,
				))
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
