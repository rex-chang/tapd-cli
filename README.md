# TAPD CLI

基于终端用户界面 (TUI) 的 TAPD 命令行工具，旨在提高开发者在终端下查看和管理 TAPD 需求与缺陷的效率。目前已集成了 AI 助手（支持 Claude 等），提供对 TAPD 数据的智能总结与辅助功能。

## 最新项目进度 🚀

截至目前，项目完成了 **AI 友好化 (AI-Friendly) 的基础改造**，主要涵盖以下已实现功能：

1. **AI 预备架构**：
   - 更新了 `Config`，支持读取 `ai.provider`、`ai.api_key` 等配置（支持从 `~/.tapd-cli/config.yaml` 灵活加载）。
   - 定义了抽象的 `ai.Provider` 接口，为后面对接多模型打下基础。
2. **Claude 模型接入**：
   - 实现了基于 Anthropic API 的 Claude Provider。
3. **分屏 TUI 交互界面**：
   - 构建了**数据区（上）+ 对话区（下）**的 Bubble Tea 终端分屏布局。
   - 实现 `Tab` 键快速切换数据浏览焦点和 AI 聊天焦点。
   - AI 对话框支持自然语言交互。
4. **快捷内置指令**：
   - `/summarize` 或 `/s`：智能总结当前所选的 TAPD 需求/缺陷。
   - `/search <query>`：将自然语言转换为 TAPD 搜索条件。
   - `/draft` 或 `/d`：一键起草专业的需求或缺陷报告格式。
   - `/help`：获取 AI 助手使用帮助。
5. **基础设施与文档设计**：
   - 补充完善了 `docs/architecture.md` 以及相关的 agent 操作规范，完成了基础架构和设计文档留档。

更详细的规划文档可见 `docs/plans/2026-02-27-ai-friendly-implementation.md`。

## 技术选型

- **语言**：Golang (1.25+)
- **TUI 框架**：[Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **TUI 组件**：[Bubbles](https://github.com/charmbracelet/bubbles)
- **样式渲染**：[Lip Gloss](https://github.com/charmbracelet/lipgloss)

## 配置示例

在 `~/.tapd-cli/config.yaml` 中准备如下配置：

```yaml
api_user: your_tapd_user
api_token: your_tapd_token
workspace_id: "your_workspace_id"

ai:
  provider: claude
  api_key: sk-ant-xxxxx
  model: claude-3-5-sonnet-20241022
```

## 编译运行

```bash
# 执行测试
go test ./... -v

# 编译代码
go build -o bin/tapd-cli ./cmd/tapd-cli

# 运行程序
./bin/tapd-cli
```
