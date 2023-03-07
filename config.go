package gone

import (
	"encoding/json"
	"os"
)

// 默认配置文件
const DEFAULT_CONFIG_PATH = "application.json"

// IConfig 配置文件接口
type IConfig interface {
	ParseFile(path string) error
}

var _ IConfig = (*AppConfig)(nil)

// AppConfig 项目配置文件
type AppConfig struct {
	Name       string `json:"name"`
	Port       int    `json:"port"`
	Production bool   `json:"production"`
}

// newAppConfig 获取带默认参数的项目配置
func newAppConfig() *AppConfig {
	return &AppConfig{
		Name:       "GONE_APP",
		Port:       9999,
		Production: false,
	}
}

// ParseFile 从文件中读取配置文件
func (config *AppConfig) ParseFile(path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(file, config); err != nil {
		return err
	}
	return nil
}
