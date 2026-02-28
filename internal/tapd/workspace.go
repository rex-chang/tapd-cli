package tapd

import (
	"encoding/json"
	"fmt"
)

// WorkspaceProject 代表 TAPD 返回的一个项目 (由于 TAPD 返回结构的特殊性，封装在 Workspace 对象中)
type WorkspaceProject struct {
	Workspace struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Desc string `json:"description"`
	} `json:"Workspace"`
}

type WorkspacesResponse struct {
	Status int                `json:"status"`
	Data   []WorkspaceProject `json:"data"`
	Info   string             `json:"info"`
}

// GetWorkspaces 获取当前用户参与的项目列表
func (c *Client) GetWorkspaces() ([]WorkspaceProject, error) {
	// TAPD 获取项目列表 API
	path := "/workspaces/projects"

	body, err := c.DoGet(path, nil)
	if err != nil {
		return nil, err
	}

	var resp WorkspacesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("解析项目列表失败: %w", err)
	}

	// 简单的错误处理
	if resp.Status != 1 {
		return nil, fmt.Errorf("TAPD 返回错误: %s", resp.Info)
	}

	return resp.Data, nil
}
