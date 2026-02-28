# TAPD CLI 开发上下文

## 1. 项目概述
本项目旨在开发一个基于终端用户界面 (TUI) 的 TAPD 命令行工具（`tapd-cli`），以提高开发者在终端下查看和管理 TAPD 需求与缺陷的效率。

## 2. 技术栈
- **语言**：Golang
- **TUI 框架**：[Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **TUI 组件库**：[Bubbles](https://github.com/charmbracelet/bubbles)
- **样式库**：[Lip Gloss](https://github.com/charmbracelet/lipgloss)

## 3. 设计原则
- **增量式进展**：从小功能（如仅展示项目列表）开始，逐步迭代。
- **清晰的意图**：代码结构保持简单直接，避免过度设计。
- **实用主义**：优先满足最常用的查询需求。

## 4. MVP (v0.1) 功能定义
1. **配置文件读取**：支持从本地（如 `~/.tapd_cli.yaml`）读取 API 凭证 (API User, API Password, Workspace ID)。
2. **基础 TUI 界面**：使用 Bubble Tea 搭建基本的全屏终端界面，具备退出机制（`q` 或 `ctrl+c`）。
3. **Mock 数据展示**：在正式接入 TAPD API 前，先使用 mock 数据验证界面的列表/表格展示功能。
4. **项目概览**：实现一个简单的视图，展示当前关注的几个重点项目状态。

## 5. 项目结构规划
```text
.
├── cmd
│   └── tapd-cli
│       └── main.go       # 程序入口
├── internal
│   ├── config            # 配置读取模块
│   ├── tapd              # TAPD API 交互模块
│   └── tui               # Bubble Tea 界面相关模型和视图
├── pkg                   # 公共工具类 (可选)
├── docs                  # 说明文档
├── go.mod
└── go.sum
```
