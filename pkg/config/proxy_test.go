package config

import (
	"testing"

	"gopkg.in/yaml.v2"
)

func TestProxyConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		expected *ProxyConfig
	}{
		{
			name: "完整代理配置",
			config: `
project:
  name: "test"
  version: "1.0.0"
proxy:
  http: "http://proxy.example.com:8080"
  https: "https://proxy.example.com:8080"
  no_proxy:
    - "localhost"
    - "127.0.0.1"
`,
			expected: &ProxyConfig{
				HTTP:  "http://proxy.example.com:8080",
				HTTPS: "https://proxy.example.com:8080",
				NoProxy: []string{
					"localhost",
					"127.0.0.1",
				},
			},
		},
		{
			name: "仅 HTTP 代理",
			config: `
project:
  name: "test"
  version: "1.0.0"
proxy:
  http: "http://proxy.example.com:8080"
`,
			expected: &ProxyConfig{
				HTTP:    "http://proxy.example.com:8080",
				HTTPS:   "",
				NoProxy: nil,
			},
		},
		{
			name: "无代理配置",
			config: `
project:
  name: "test"
  version: "1.0.0"
`,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewConfigLoader(".")
			config, err := loader.LoadFromString(tt.config)
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			if tt.expected == nil {
				if config.Proxy != nil {
					t.Errorf("Expected nil proxy, got %+v", config.Proxy)
				}
				return
			}

			if config.Proxy == nil {
				t.Errorf("Expected proxy config, got nil")
				return
			}

			if config.Proxy.HTTP != tt.expected.HTTP {
				t.Errorf("Expected HTTP proxy %s, got %s", tt.expected.HTTP, config.Proxy.HTTP)
			}

			if config.Proxy.HTTPS != tt.expected.HTTPS {
				t.Errorf("Expected HTTPS proxy %s, got %s", tt.expected.HTTPS, config.Proxy.HTTPS)
			}

			if len(config.Proxy.NoProxy) != len(tt.expected.NoProxy) {
				t.Errorf("Expected %d no_proxy entries, got %d", len(tt.expected.NoProxy), len(config.Proxy.NoProxy))
				return
			}

			for i, expected := range tt.expected.NoProxy {
				if i >= len(config.Proxy.NoProxy) || config.Proxy.NoProxy[i] != expected {
					t.Errorf("Expected no_proxy[%d] = %s, got %s", i, expected, config.Proxy.NoProxy[i])
				}
			}
		})
	}
}

// LoadFromString 从字符串加载配置（测试辅助函数）
func (cl *ConfigLoader) LoadFromString(configStr string) (*ProjectConfig, error) {
	var config ProjectConfig
	if err := yaml.Unmarshal([]byte(configStr), &config); err != nil {
		return nil, err
	}

	// 设置依赖项名称
	for name, dep := range config.Dependencies {
		dep.Name = name
		config.Dependencies[name] = dep
	}
	config.ProjectRoot = cl.baseDir

	return &config, nil
}

func TestProxyConfigMerge(t *testing.T) {
	globalConfig := `
project:
  name: "global"
  version: "1.0.0"
proxy:
  http: "http://global.proxy.com:8080"
  https: "https://global.proxy.com:8080"
  no_proxy:
    - "global.local"
`

	localConfig := `
project:
  name: "local"
  version: "1.0.0"
proxy:
  http: "http://local.proxy.com:8080"
  no_proxy:
    - "local.local"
    - "localhost"
`

	loader := NewConfigLoader(".")

	// 模拟合并配置
	global, _ := loader.LoadFromString(globalConfig)
	local, _ := loader.LoadFromString(localConfig)

	merged := loader.mergeConfigs(global, local)

	// 验证本地配置覆盖全局配置
	if merged.Proxy.HTTP != "http://local.proxy.com:8080" {
		t.Errorf("Expected HTTP proxy from local config, got %s", merged.Proxy.HTTP)
	}

	if merged.Proxy.HTTPS != "" {
		t.Errorf("Expected empty HTTPS proxy (not set in local), got %s", merged.Proxy.HTTPS)
	}

	if len(merged.Proxy.NoProxy) != 2 {
		t.Errorf("Expected 2 no_proxy entries, got %d", len(merged.Proxy.NoProxy))
	}

	if merged.Proxy.NoProxy[0] != "local.local" {
		t.Errorf("Expected first no_proxy entry from local config, got %s", merged.Proxy.NoProxy[0])
	}
}
