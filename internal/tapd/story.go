package tapd

import (
	"encoding/json"
	"fmt"
)

// Story 需求基本字段
type Story struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Creator     string `json:"creator"`
	Status      string `json:"status"`
}

// StoryItem TAPD 返回的具体需求条目包装
type StoryItem struct {
	Story Story `json:"Story"`
}

// StoriesResponse 需求列表响应
type StoriesResponse struct {
	Status int         `json:"status"`
	Data   []StoryItem `json:"data"`
	Info   string      `json:"info"`
}

// GetStories 获取当前工作区下的需求列表
func (c *Client) GetStories() ([]StoryItem, error) {
	path := "/stories"
	// 默认参数：限制返回字段，减少数据量；添加 status 筛选以避免返回太多关闭的故事 (可视情况拓展)
	query := map[string]string{
		"limit":  "30",
		"fields": "id,name,description,creator,status",
	}

	body, err := c.DoGet(path, query)
	if err != nil {
		return nil, err
	}

	var resp StoriesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("解析需求列表失败: %w", err)
	}

	if resp.Status != 1 {
		return nil, fmt.Errorf("TAPD 返回错误: %s", resp.Info)
	}

	return resp.Data, nil
}
