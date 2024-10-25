package main

import (
	"embed"
	"os"
	"os/signal"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/samber/lo"

	"sealchat/api"
	"sealchat/model"
	"sealchat/pm"
	"sealchat/utils"
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

	lo.Must0(os.MkdirAll("./data", 0755))
	config := utils.ReadConfig()

	model.DBInit(config.DSN)
	cleanUp := func() {
		if db := model.GetDB(); db != nil {
			if sqlDB, err := db.DB(); err == nil {
				_ = sqlDB.Close()
			}
		}
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		_ = <-c
		cleanUp()
		os.Exit(0)
	}()

	pm.Init()
	// model.UserRoleMappingCreate(&model.UserRoleMappingModel{
	// 	UserID:   "e6ww4e_jUU9LXOKjo3h3z",
	// 	RoleID:   "sys-admin",
	// 	RoleType: "system",
	// })

	// perm := pm.GetAllSysPermByUid("e6ww4e_jUU9LXOKjo3h3z")
	// x, err := json.Marshal(perm)
	// fmt.Println(string((x)), err)

	autoSave := func() {
		t := time.NewTicker(3 * 60 * time.Second)
		for {
			<-t.C
			model.FlushWAL()
		}
	}
	go autoSave()

	api.Init(config, embedDirStatic)
}
