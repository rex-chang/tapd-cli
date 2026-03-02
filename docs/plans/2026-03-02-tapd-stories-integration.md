# TAPD 需求列表接入与渲染支持

## 目标
当前 TUI 默认展示用户参与且由于历史原因仅有项目 (Workspaces) 列表展示。在实际工作流中，已经配置了特定的 Workspace ID 后，用户打开 CLI 工具希望优先看到当前迭代下或近期的 "需求 (Stories)"。
本阶段任务：连接 TAPD 的需求接口 (`/stories`)，并调整 TUI 应用在启动后直接显示需求列表。

## 计划拆解

### Task 1: 增加 tapd.GetStories 接口
- **路径**: `/stories`
- **请求参数**: `workspace_id` (全局拦截器已处理), 需要限制返回字段或数量防止慢查询。
- **返回结构**: 支持反序列化 TAPD "/stories" 响应结构，例如：
  ```json
  {
      "status": 1,
      "data": [
          {
              "Story": {
                  "id": "11xxxxxx",
                  "name": "需求标题",
                  "status": "status_1",
                  "creator": "创建人"
              }
          }
      ]
  }
  ```
- **输出**: `internal/tapd/story.go` 以及 `internal/tapd/story_test.go`。

### Task 2: 将 TUI 默认视图改为"需求列表"
- **描述**: 修改 `internal/tui/app.go`
- **步骤**:
  1. 将初始化时的 `fetchWorkspaces` 替换为 `fetchStories` （请求该接口加载故事）。
  2. 新增或修改 `storiesLoadedMsg`。
  3. 修改 `Model.Update()` 对列表模型注入需求 (Item title = Story.Name, item desc = Status / Creator)。
  4. 测试 TUI 的显示和切换能否正常工作。

### Task 3: 优化 AI Prompt Injection 
让 TAPD 需求的内容 (`name` + `description` 如果有，或者目前只取 `name` 和状态) 能够更详尽地传递给 AI Prompt (`buildContextMessages`)，增强 `/summarize` 时的信息准确性。
