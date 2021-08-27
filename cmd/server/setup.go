// 命令行配置
package main

import (
	"context"
	"errors"
	"myclush/service"
	"myclush/utils"
	"os"
	"sort"

	"myclush/global"
	log "myclush/logger"

	"github.com/urfave/cli/v2"
)

// 全局变量
var (
	// 指定agent节点列表
	nodes string
	// 调试模式
	debug bool
	// 端口
	port int
)

// 全局选项参数配置
var (
	globalFlagForDebug *cli.BoolFlag = &cli.BoolFlag{
		Name:        "debug",
		Aliases:     []string{"d"},
		Value:       false,
		Usage:       "set log level debug",
		Destination: &debug,
	}
	globalFlagForPort *cli.IntFlag = &cli.IntFlag{
		Name:        "port",
		Aliases:     []string{"p"},
		Value:       1995,
		Usage:       "grpc service port",
		Destination: &port,
	}
)

func run(ctx context.Context, cancel context.CancelFunc) error {
	app := &cli.App{
		// 基本信息
		// Name:     name,
		// HelpName: name,
		Version: global.Version,
		// Description: descriptions,
		Usage: global.Descriptions,
		// 子命令执行前的设置
		Before: Before,
		Authors: []*cli.Author{
			{
				Name:  global.Author,
				Email: global.Email,
			},
		},
		// 全局选项参数配置
		Flags: []cli.Flag{
			globalFlagForDebug,
			globalFlagForPort,
		},
		Action: func(c *cli.Context) error {
			service.PutStreamServerServiceSetup(ctx, cancel, c.App.Name, port)
			return nil
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	return err
}

func setLogLevel(debug bool) {
	if debug {
		log.SetLevel(log.DEBUG)
		log.Debug("Logger Setup In DEBUG Mode")
	} else {
		log.SetSilent()
	}
}

func Before(c *cli.Context) error {
	setLogLevel(c.Bool("debug"))
	uid, gid, err := utils.UserInfo()
	if err != nil {
		return err
	}
	if uid != "0" && gid != "0" {
		return errors.New("permission denied, need root privileges")
	}
	// gen tls
	clientCreds, err := service.GenClientTransportCredentials()
	if err != nil {
		return err
	}
	global.ClientTransportCredentials = clientCreds
	creds, err := service.GenServerTransportCredentials()
	if err != nil {
		return err
	}
	global.ServerTransportCredentials = creds
	return nil
}
