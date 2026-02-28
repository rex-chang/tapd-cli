package tapd

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rex-chang/tapd-cli/internal/config"
)

const BaseURL = "https://api.tapd.cn"

type Client struct {
	cfg        *config.Config
	httpClient *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) DoGet(path string, query map[string]string) ([]byte, error) {
	reqURL := BaseURL + path
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	for k, v := range query {
		q.Add(k, v)
	}

	// 如果全局配置了 workspace_id，则默认附带上。当然也可以由不同的 API 决定
	if c.cfg.WorkspaceID != "" && q.Get("workspace_id") == "" {
		q.Add("workspace_id", c.cfg.WorkspaceID)
	}
	req.URL.RawQuery = q.Encode()

	// 使用 Basic Auth，优先尝试 API Token
	authPassword := c.cfg.APIToken
	if authPassword == "" {
		authPassword = c.cfg.APIPassword
	}
	req.SetBasicAuth(c.cfg.APIUser, authPassword)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body) // 读取错误详情以便排查
		return nil, fmt.Errorf("API 访问失败: %s (HTTP %d)", string(body), resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
