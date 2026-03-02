package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// AIConfig 存储 AI Provider 相关配置
type AIConfig struct {
	Provider string `yaml:"provider"` // claude | openai
	APIKey   string `yaml:"api_key"`
	Model    string `yaml:"model"`
}

// Config 存储了 TAPD 相关的凭据和基础配置
type Config struct {
	AccessToken string   `yaml:"access_token"` // 个人访问令牌
	WorkspaceID string   `yaml:"workspace_id"`
	AI          AIConfig `yaml:"ai"`
}

// LoadConfigFromPath 从指定路径加载配置（方便测试）
func LoadConfigFromPath(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	if cfg.AccessToken == "" || cfg.WorkspaceID == "" {
		return nil, fmt.Errorf("配置文件中必需包含 access_token 和 workspace_id")
	}

	return &cfg, nil
}

// LoadConfig 从 ~/.tapd-cli/config.yaml 加载配置
func LoadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("无法获取用户主目录: %w", err)
	}

	configDir := filepath.Join(homeDir, ".tapd-cli")
	configPath := filepath.Join(configDir, "config.yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := promptConfig(configDir, configPath); err != nil {
			return nil, fmt.Errorf("初始化配置失败: %w", err)
		}
	}

	return LoadConfigFromPath(configPath)
}

// promptConfig 交互式引导用户初始化配置
func promptConfig(configDir, configPath string) error {
	fmt.Println("未找到配置文件，开始初始化 TAPD CLI 配置...")
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("TAPD Access Token: ")
	accessToken, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("读取 Access Token 失败: %w", err)
	}
	accessToken = strings.TrimSpace(accessToken)

	fmt.Print("TAPD Workspace ID: ")
	workspaceID, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("读取 Workspace ID 失败: %w", err)
	}
	workspaceID = strings.TrimSpace(workspaceID)

	if accessToken == "" || workspaceID == "" {
		return fmt.Errorf("必需字段不可为空")
	}

	cfg := Config{
		AccessToken: accessToken,
		WorkspaceID: workspaceID,
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	fmt.Printf("\n配置已成功保存至 %s\n\n", configPath)
	return nil
}
