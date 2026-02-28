# tapd-cli 架构文档

## 模块依赖图

```
main.go
  → config.LoadConfig()        # 加载配置，含可选 AI 配置
  → ai.Provider (可选)         # 根据配置初始化 Claude/OpenAI Provider
  → tui.InitialModel(cfg, p)   # 注入配置和 Provider，创建 TUI 模型
  → tea.NewProgram(model)      # 启动 Bubble Tea 事件循环
```

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

```
用户按 Enter
  → chat.Update() 产生 sendMessageCmd
  → app.Update() 拦截，调用 callAICmd(provider, messages)
  → AI Provider.Chat() 异步执行
  → 返回 aiCallResultMsg
  → app.Update() 转发给 chat.Update(aiResponseMsg)
  → chat.refreshViewport() 更新显示
```

## 分屏布局

```
┌─────────────────────────────────────┐
│   TAPD 数据区（上方 ~70%）            │
│   - 项目列表 / 需求详情 / 缺陷列表    │
│   - 支持上下滚动                      │
├─────────────────────────────────────┤
│   AI 对话区（下方 ~30%）              │
│   - 消息历史显示（viewport）          │
│   - 输入框（textarea）                │
│   - Tab 切换焦点                      │
└─────────────────────────────────────┘
```

## AI 内置指令

| 指令 | 功能 |
|------|------|
| /summarize 或 /s | 总结当前条目 |
| /search <query> | 自然语言转搜索条件 |
| /draft 或 /d | 起草需求/缺陷报告 |
| /help | 列出可用功能 |
