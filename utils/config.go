package utils

import (
	"fmt"
	"github.com/knadh/koanf/providers/file"
	"os"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/samber/lo"
)

type AppConfig struct {
	ServeAt                   string `json:"serveAt" yaml:"serveAt"`
	Domain                    string `json:"domain" yaml:"domain"`
	RegisterOpen              bool   `json:"registerOpen" yaml:"registerOpen"`
	WebUrl                    string `json:"webUrl" yaml:"webUrl"`
	ChatHistoryPersistentDays int64  `json:"chatHistoryPersistentDays" yaml:"chatHistoryPersistentDays"`
	ImageSizeLimit            int64  `json:"imageSizeLimit" yaml:"imageSizeLimit"` // in kb
	ImageCompress             bool   `json:"imageCompress" yaml:"imageCompress"`
	DSN                       string `json:"-" yaml:"dbUrl"`
}

// 注: 实验型使用koanf，其实从需求上讲目前并无必要
var k = koanf.New(".")

func ReadConfig() *AppConfig {
	config := AppConfig{
		ServeAt:                   ":3212",
		Domain:                    "127.0.0.1:3212",
		RegisterOpen:              true,
		WebUrl:                    "/",
		ChatHistoryPersistentDays: 60,
		ImageSizeLimit:            8192,
		ImageCompress:             true,
		DSN:                       "./data/chat.db",
	}

	lo.Must0(k.Load(structs.Provider(&config, "yaml"), nil))

	isNotExist := false
	if err := k.Load(file.Provider("config.yaml"), yaml.Parser()); err != nil {
		fmt.Printf("配置读取失败: %v\n", err)

		if os.IsNotExist(err) {
			isNotExist = true
		} else {
			os.Exit(-1)
		}
	}

	if isNotExist {
		WriteConfig(nil)
	} else {
		if err := k.Unmarshal("", &config); err != nil {
			fmt.Printf("配置解析失败: %v\n", err)
			os.Exit(-1)
		}
	}

	k.Print()
	return &config
}

func WriteConfig(config *AppConfig) {
	if config != nil {
		err := k.Load(structs.Provider(&config, "yaml"), nil)
		if err != nil {
			fmt.Println("错误: 配置文件序列化失败")
			return
		}
	}

	content, err := yaml.Parser().Marshal(k.Raw())
	if err != nil {
		fmt.Println("错误: 配置文件序列化失败")
		return
	}
	err = os.WriteFile("./config.yaml", content, 0644)
	if err != nil {
		fmt.Println("错误: 配置文件写入失败")
	}
}
