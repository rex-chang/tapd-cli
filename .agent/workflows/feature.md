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
