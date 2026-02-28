package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rex-chang/tapd-cli/internal/ai"
	"github.com/rex-chang/tapd-cli/internal/ai/claude"
	"github.com/rex-chang/tapd-cli/internal/config"
	"github.com/rex-chang/tapd-cli/internal/tui"
)

func main() {
	// 1. 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("配置错误: %v\n", err)
		os.Exit(1)
	}

	// 2. 根据配置初始化 AI Provider（可选）
	var provider ai.Provider
	if cfg.AI.APIKey != "" {
		switch cfg.AI.Provider {
		case "claude", "": // 默认使用 Claude
			model := cfg.AI.Model
			if model == "" {
				model = "claude-3-5-sonnet-20241022"
			}
			provider = claude.NewProvider(cfg.AI.APIKey, model)
		}
	}

	// 3. 初始化 TUI 并启动
	model := tui.InitialModel(cfg, provider)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("启动 tapd-cli 失败: %v\n", err)
		os.Exit(1)
	}
}
