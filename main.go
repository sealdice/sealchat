package main

import (
	"embed"
	"github.com/samber/lo"
	"os"
	"sealchat/api"
	"sealchat/model"
	"sealchat/utils"

	"github.com/jessevdk/go-flags"
)

//go:embed ui/dist
var embedDirStatic embed.FS

func main() {
	var opts struct {
		Install   bool `short:"i" long:"install" description:"安装为系统服务"`
		Uninstall bool `long:"uninstall" description:"删除系统服务"`
	}
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		return
	}

	if opts.Install {
		serviceInstall(true)
		return
	}

	if opts.Uninstall {
		serviceInstall(false)
		return
	}

	lo.Must0(os.MkdirAll("./data", 0644))
	config := utils.ReadConfig()

	model.DBInit()
	api.Init(config, embedDirStatic)
}
