# 从 API User/Password 迁移到个人访问令牌

**日期**: 2026-03-02
**类型**: Breaking Change
**状态**: 计划中

## 背景

当前 `tapd-cli` 支持三种认证方式：
- API User + API Password（Basic Auth）
- API User + API Token（Basic Auth 变体）
- 个人访问令牌（未正确实现）

根据 TAPD 官方文档，个人访问令牌是推荐的独立认证方式，应使用 Bearer Token 格式。

## 目标

- 移除 API User/Password 支持
- 使用纯个人访问令牌 + Bearer Token 认证
- 简化配置和代码逻辑

## 设计方案

### 配置文件变更

**旧配置**:
```yaml
api_user: "your_api_user"
api_password: "your_api_password"
api_token: "your_token"
workspace_id: "your_workspace_id"
```

**新配置**:
```yaml
access_token: "your_personal_access_token"
workspace_id: "your_workspace_id"
```

### 代码变更

#### 1. `internal/config/config.go`
- 移除字段: `APIUser`, `APIPassword`
- 重命名字段: `APIToken` → `AccessToken`
- 简化验证: 只检查 `AccessToken` 和 `WorkspaceID`

#### 2. `internal/tapd/client.go`
- 认证方式: `req.SetBasicAuth()` → `req.Header.Set("Authorization", "Bearer "+token)`

#### 3. 交互式配置
- 只提示: Access Token + Workspace ID

### 迁移指南

用户需要：
1. 访问 TAPD → 我的设置 → 个人访问令牌
2. 创建令牌并保存
3. 更新配置文件

## 影响范围

### 修改文件
- `internal/config/config.go`
- `internal/tapd/client.go`
- `README.md`
- 所有测试文件

### 破坏性变更
- 旧配置文件失效
- 用户必须手动迁移

## 实施步骤

见 `implementation-plan.md`
