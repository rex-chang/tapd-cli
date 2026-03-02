package tapd

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rex-chang/tapd-cli/internal/config"
)

const DefaultBaseURL = "https://api.tapd.cn"

type Client struct {
	cfg        *config.Config
	httpClient *http.Client
	baseURL    string
}

// ClientOption 定义客户端配置选项
type ClientOption func(*Client)

// WithBaseURL 用于在测试中覆盖默认的 API 基础 URL
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

func NewClient(cfg *config.Config, opts ...ClientOption) *Client {
	c := &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: DefaultBaseURL,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) DoGet(path string, query map[string]string) ([]byte, error) {
	reqURL := c.baseURL + path
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

	// 使用 Bearer Token 认证
	req.Header.Set("Authorization", "Bearer "+c.cfg.AccessToken)
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
