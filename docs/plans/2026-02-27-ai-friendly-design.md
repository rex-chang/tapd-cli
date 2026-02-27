# tapd-cli AI 友好化设计方案

日期：2026-02-27

## 背景

将 tapd-cli 改造为"AI 开发友好"项目，包含两个维度：
1. **AI 辅助开发**：让 AI 编码助手更容易理解、扩展和维护这个项目
2. **AI 功能集成**：在工具本身集成 AI 能力，提升用户使用体验

## 目标

- 上下分屏 TUI 布局（类似 Claude Code 交互方式）
- 可插拔的多 AI Provider 接口
- 三大 AI 功能：内容摘要、智能搜索、写作辅助
- 自动感知当前上下文 + 支持手动 @引用
- 完善开发者体验基础设施（文档、注释规范、Agent 工作流）

## 整体架构

### 维度一：AI 开发基础设施

```
docs/
  context.md          # 持续维护的项目上下文
  architecture.md     # 模块关系、设计决策
  plans/              # 功能设计文档（按日期归档）
.agent/
  workflows/          # 标准开发流程定义
CLAUDE.md             # Claude Code 指引，持续更新
```

代码规范：
- 统一中文注释
- 接口优先设计（便于 AI 生成 mock 和测试）
- 明确模块职责边界

### 维度二：AI 功能集成

#### 新增模块

```
internal/
  ai/
    provider.go       # Provider 抽象接口
    claude/           # Claude API 实现
    openai/           # OpenAI API 实现
    context.go        # 上下文感知逻辑（当前 TAPD 数据注入）
  tui/
    app.go            # 现有，扩展分屏布局
    chat.go           # 新增：底部 AI 对话框组件
```

#### AI Provider 接口设计

```go
type Provider interface {
    // Chat 发送消息并获取完整响应
    Chat(ctx context.Context, messages []Message) (string, error)
    // Stream 发送消息并流式返回响应
    Stream(ctx context.Context, messages []Message) (<-chan string, error)
    // Name 返回 Provider 名称
    Name() string
}

type Message struct {
    Role    string // "user" | "assistant" | "system"
    Content string
}
```

#### TUI 布局设计

```
┌─────────────────────────────────────┐
│                                     │
│   TAPD 数据区（上方 ~70%）            │
│   - 项目列表 / 需求详情 / 缺陷列表    │
│   - 支持上下滚动                      │
│                                     │
├─────────────────────────────────────┤
│   AI 对话区（下方 ~30%）              │
│   > 输入框（Tab 键切换焦点）           │
│   AI: 这个需求的核心是...             │
│                                     │
└─────────────────────────────────────┘
```

**快捷键设计**：
- `Tab`：切换上下区域焦点
- `/`：快速唤起 AI 对话框并聚焦
- `@`：在对话框中触发上下文引用补全

#### 上下文感知机制

1. **自动感知**：AI 系统 Prompt 中自动注入当前光标所在 TAPD 条目的完整信息
2. **手动引用**：在对话框输入 `@` 触发补全菜单，选择引用当前/其他条目

#### 三大 AI 功能

| 功能 | 触发方式 | 描述 |
|------|----------|------|
| 内容摘要 | `/summarize` 或 `s` 快捷键 | 总结当前需求/缺陷的核心信息 |
| 智能搜索 | `/search <自然语言>` | 将自然语言转换为 TAPD 查询条件 |
| 写作辅助 | `/draft` | 起草需求描述、缺陷报告、评论 |

## 配置扩展

`~/.tapd_cli.yaml` 新增 AI 配置节：

```yaml
api_user: "your_api_user"
api_password: "your_api_password"
workspace_id: "your_workspace_id"

# 新增 AI 配置
ai:
  provider: "claude"      # claude | openai
  api_key: "sk-..."
  model: "claude-3-5-sonnet-20241022"
```

## 实现优先级

1. **P0**：AI Provider 接口 + Claude 实现（基础能力）
2. **P0**：TUI 分屏布局 + 对话框组件
3. **P1**：上下文自动感知注入
4. **P1**：内容摘要功能
5. **P2**：智能搜索、写作辅助
6. **P2**：OpenAI Provider 实现
7. **P3**：AI 开发基础设施完善
