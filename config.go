package gone

import (
	"encoding/json"
	"os"
)

// IFileConfig 文件配置
type IFileConfig interface {
	ParseFile(name string) error
	ParseContent(data []byte) error
	IsValid(name string) bool
}

// AppConfig 项目配置
type AppConfig struct {
	Name       string `json:"name"`
	Port       int    `json:"port"`
	Production bool   `json:"production"`
	MainFile   string `json:"mainFile"`
	Loggers    []struct {
		Name   string `json:"name"`
		Output bool   `json:"output"`
	} `json:"loggers"`
	Watcher struct {
		Includes []string `json:"includes"`
		Excludes []string `json:"excludes"`
	} `json:"watcher"`
}

// newAppConfig 构造默认项目配置
func newAppConfig() *AppConfig {
	return &AppConfig{
		Name:       "GONE",
		Port:       9999,
		Production: false,
		MainFile:   "main.go",
	}
}

// IsValid 判断文件是否可用
func (config *AppConfig) IsValid(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

// ParseFile 从文件读取配置
func (config *AppConfig) ParseFile(name string) error {
	file, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, &config)
}

// ParseContent 从二进制加载配置
func (config *AppConfig) ParseContent(data []byte) error {
	return json.Unmarshal(data, &config)
}
