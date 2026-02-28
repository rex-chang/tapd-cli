# AGENTS.md

本文件为 AI 编程助手（Agent）提供在 tapd-cli 代码库中工作的规范指南。

## 项目概述

tapd-cli 是一个基于终端用户界面 (TUI) 的 TAPD（腾讯敏捷产品开发平台）命令行工具，使用 Go 开发。采用 Bubble Tea 框架构建全屏终端界面，允许开发者在终端直接查看和管理 TAPD 需求与缺陷。

## 常用命令

### 构建
```bash
# 构建二进制到 bin/ 目录
go build -o bin/tapd-cli ./cmd/tapd-cli

# 直接运行（需要已配置 ~/.tapd_cli.yaml）
go run ./cmd/tapd-cli/main.go
```

### 测试
```bash
# 运行所有测试
go test ./...

# 运行单个测试文件
go test -v ./internal/config

# 运行单个测试函数
go test -v -run TestFunctionName ./package/path

# 运行带覆盖率报告的测试
go test -cover ./...
```

### 代码质量
```bash
# 代码格式化
go fmt ./...

# 静态检查
go vet ./...

# 下载/整理依赖
go mod tidy

# 下载依赖
go mod download
```

## 代码架构

项目采用标准 Go 项目布局，严格分层：

```
cmd/tapd-cli/main.go      # 程序入口
internal/config/           # 配置层：从 ~/.tapd-cli/config.yaml 读取并验证凭证
internal/tapd/             # API 层：TAPD HTTP 客户端 + 数据模型
internal/tui/              # UI 层：Bubble Tea MUV 模式的全屏 TUI 应用
pkg/                       # 公共工具（预留）
```

分层依赖关系：`main.go → config → tapd.Client → tui.Model`

## 代码风格规范

### 包组织
- 标准库导入在前，第三方库居中，本项目包在后，每组之间空一行
- 包名使用小写单数形式，避免下划线
- internal 包用于内部实现，pkg 用于可复用的公共代码

### 命名规范
- **文件命名**：小写，使用下划线分隔（如 `workspace.go`）
- **类型/结构体**：PascalCase，首字母大写表示导出
- **接口**：以 `-er` 结尾（如 `Reader`, `Writer`）
- **函数/方法**：PascalCase 导出，camelCase 私有
- **变量**：camelCase，简短但有含义
- **常量**：PascalCase 或全大写下划线（如 `BaseURL`）
- **错误变量**：以 `Err` 开头（如 `ErrNotFound`）

### 类型定义
- 使用结构体标签定义 JSON/YAML 字段映射
- 嵌套结构体用于处理 API 返回的特殊嵌套格式
- 状态机使用 iota 定义枚举常量

```go
type viewState int

const (
    stateLoading viewState = iota
    stateList
)
```

### 错误处理
- 使用 `fmt.Errorf` 包装错误，添加上下文信息
- 使用 `%w` 动词保留原始错误，支持错误链
- 使用 `errors.Is` 进行错误类型判断
- 错误信息使用中文，方便用户理解

```go
if err != nil {
    return nil, fmt.Errorf("操作失败: %w", err)
}
```

### 注释规范
- 所有导出的类型、函数、变量必须添加注释
- 注释以被注释对象的名称开头
- 使用中文注释说明功能意图

```go
// LoadConfig 从 ~/.tapd-cli/config.yaml 加载配置
func LoadConfig() (*Config, error) { ... }
```

### 配置与常量
- 外部 API URL 使用 const 定义
- HTTP 超时使用 time.Duration
- 敏感配置文件使用 0600 权限

### TUI 开发规范
- 使用 Bubble Tea 的 Model-Update-View 模式
- 自定义消息类型用于异步回调（如 `workspacesLoadedMsg`）
- 样式使用 Lip Gloss 定义，集中管理在 var 块中
- 退出机制统一使用 `q` 或 `ctrl+c`

## TAPD API 集成

- Base URL: `https://api.tapd.cn`
- 认证：HTTP Basic Auth（优先使用 API Token）
- `workspace_id` 自动附加到所有请求参数
- 请求超时：10 秒

## 新增功能指引

- 新增 TAPD API 接口 → `internal/tapd/` 下添加新文件，参考 `workspace.go` 结构
- 新增 TUI 视图 → `internal/tui/` 下添加，通过新增 `viewState` 常量切换
- 新增配置项 → `internal/config/config.go` 的 Config 结构体

## 外部资源

- [CLAUDE.md](./CLAUDE.md) - 详细的项目上下文和架构说明
- [docs/context.md](./docs/context.md) - 设计原则和 MVP 功能定义
- [TAPD API 文档](https://www.tapd.cn/help/show#1120001721001000093)

## 相关规则文件

**注意**：当前项目未配置 Cursor Rules 或 GitHub Copilot Instructions。
如需添加，请创建：
- `.cursorrules` 文件，或
- `.cursor/rules/` 目录下的规则文件，或
- `.github/copilot-instructions.md` 文件
