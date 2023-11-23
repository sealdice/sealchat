package main

import (
	"fmt"
	"os"
	"sealchat/api"
	"sealchat/model"

	"github.com/jessevdk/go-flags"
	"github.com/spf13/afero"
)

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

	var AppFs = afero.NewMemMapFs()
	fmt.Println("111", AppFs)

	model.DBInit()
	api.Init()
}
