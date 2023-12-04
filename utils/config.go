package utils

import (
	"fmt"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
	"os"
)

type AppConfig struct {
	ServeAt                   string `json:"serveAt" yaml:"serveAt"`
	Domain                    string `json:"domain" yaml:"domain"`
	RegisterOpen              bool   `json:"registerOpen" yaml:"registerOpen"`
	WebUrl                    string `json:"webUrl" yaml:"webUrl"`
	ChatHistoryPersistentDays int64  `json:"chatHistoryPersistentDays" yaml:"chatHistoryPersistentDays"`
	ImageSizeLimit            int64  `json:"imageSizeLimit" yaml:"imageSizeLimit"` // in kb
	ImageCompress             bool   `json:"imageCompress" yaml:"imageCompress"`
}

func ReadConfig() *AppConfig {
	config := AppConfig{
		ServeAt:                   ":3212",
		Domain:                    "127.0.0.1:3212",
		RegisterOpen:              true,
		WebUrl:                    "/",
		ChatHistoryPersistentDays: 60,
		ImageSizeLimit:            4096,
		ImageCompress:             true,
	}
	content, err := os.ReadFile("./config.yaml")
	if err == nil {
		lo.Must0(yaml.Unmarshal(content, &config))
	}
	content, err = yaml.Marshal(config)
	if err == nil {
		lo.Must0(os.WriteFile("./config.yaml", content, 0644))
	}
	return &config
}

func WriteConfig(config *AppConfig) {
	content, err := yaml.Marshal(config)
	if err != nil {
		fmt.Println("错误: 配置文件序列化失败")
		return
	}
	err = os.WriteFile("./config.yaml", content, 0644)
	if err != nil {
		fmt.Println("错误: 配置文件写入失败")
	}
}
