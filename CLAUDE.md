# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

`tapd-cli` 是一个基于终端用户界面 (TUI) 的 TAPD（腾讯敏捷产品开发平台）命令行工具，使用 Go 开发。允许开发者在终端直接查看和管理 TAPD 需求与缺陷。

## 常用命令

```bash
# 构建二进制到 bin/ 目录
go build -o bin/tapd-cli ./cmd/tapd-cli

# 直接运行（需要已配置 ~/.tapd_cli.yaml）
go run ./cmd/tapd-cli/main.go

# 运行测试
go test ./...

# 运行单个测试
go test ./internal/config/... -v -run TestLoadConfig_WithAI

# 代码格式化
go fmt ./...

# 静态检查
go vet ./...

# 下载依赖
go mod tidy

# 编译并运行
go build -o bin/tapd-cli ./cmd/tapd-cli && ./bin/tapd-cli
```

## 配置文件

运行前需要在 `~/.tapd_cli.yaml` 中配置：

```yaml
api_user: "your_api_user"
api_password: "your_api_password"
workspace_id: "your_workspace_id"

# 可选 AI 配置（未配置则不显示 AI 对话区）
ai:
  provider: "claude"  # 或 "openai"
  claude:
    api_key: "your_api_key"
    model: "claude-3-5-sonnet-20241022"
  openai:
    api_key: "your_api_key"
    model: "gpt-4"
```

## 代码架构

项目采用标准 Go 项目布局，严格分层：

```
cmd/tapd-cli/main.go     # 程序入口：加载配置 → 初始化 TUI Model → 启动 Bubble Tea Program
internal/config/         # 配置层：从 ~/.tapd_cli.yaml 读取并验证凭证
internal/tapd/           # API 层：TAPD HTTP 客户端（Basic Auth）+ 数据模型
internal/ai/             # AI 抽象层：Provider 接口定义
internal/ai/claude/      # AI Claude 实现
internal/tui/            # UI 层：Bubble Tea MUV 模式的全屏 TUI 应用
pkg/                     # 公共工具（预留，暂未使用）
```

### 分层依赖关系

```
main.go → config → tapd.Client → tui.Model
              ↓
          ai.Provider (可选)
              ↓
          tui.ChatModel
```

- `config.LoadConfig()` 加载配置，失败则程序退出
- `tapd.NewClient(cfg)` 注入配置，提供 `DoGet()` 通用方法和 `GetWorkspaces()` 等具体接口
- `ai.NewProvider(cfg)` 根据配置创建 Claude/OpenAI Provider（可选）
- `tui.NewModel(cfg, client, provider)` 使用 Bubble Tea 的 Model-Update-View 模式，异步拉取数据通过自定义消息类型（`workspacesLoadedMsg`、`errMsg`）回调

### TUI 状态机

`tui/app.go` 的视图状态：
- `stateLoading` → 显示 Spinner 动画，后台异步调用 API
- `stateList` → 显示项目列表，使用 `bubbles/list` 组件

### TAPD API 客户端

- Base URL: `https://api.tapd.cn`
- 认证：HTTP Basic Auth（APIUser + APIPassword）
- `workspace_id` 自动附加到所有请求参数
- 请求超时：10 秒

### AI 功能

- **上下文感知**：当前选中条目自动注入 AI 系统 Prompt
- **内置指令**：/summarize（总结）、/search（搜索）、/draft（起草）、/help（帮助）
- **对话模式**：分屏下方为 AI 对话区，支持 Enter 发送、Alt+Enter 换行
- **降级策略**：未配置 AI Provider 时显示友好提示，不影响 TAPD 功能使用

## 技术栈

| 库 | 版本 | 用途 |
|----|------|------|
| `charmbracelet/bubbletea` | v1.3.10 | TUI 框架核心 |
| `charmbracelet/bubbles` | v1.0.0 | 预制组件（List、Spinner、Textarea、Viewport） |
| `charmbracelet/lipgloss` | v1.1.0 | 终端样式渲染 |
| `gopkg.in/yaml.v3` | v3.0.1 | 配置文件解析 |

### AI 集成

项目使用可插拔 AI Provider 架构：

- **Provider 接口**：定义在 `internal/ai/provider.go`
  - `Chat(ctx, []Message) (string, error)`：同步调用，返回完整响应
  - `Stream(ctx, []Message) (<-chan StreamChunk, error)`：流式调用

- **现有实现**：
  - `internal/ai/claude/`：Anthropic Claude API 实现
  - 测试使用 `httptest.NewServer` mock，无需真实 API 密钥

- **扩展新 Provider**：
  1. 创建 `internal/ai/<name>/` 目录
  2. 实现 `Provider` 接口
  3. 编写测试（优先使用 httptest mock）
  4. 在 `main.go` 的配置初始化代码中添加新的 case 分支

## 开发约定

- 遵循 `docs/context.md` 中的设计原则：增量式进展、清晰意图、实用主义
- 新增 TAPD API 接口在 `internal/tapd/` 下添加新文件（参考 `workspace.go` 结构）
- TUI 新视图在 `internal/tui/` 下添加，通过新增 `viewState` 常量切换
- 目前无 Makefile，构建/运行使用标准 `go` 命令
