// 命令行配置
package main

import (
	"context"
	"errors"
	"myclush/service"
	"myclush/utils"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"myclush/global"
	log "myclush/logger"

	"github.com/urfave/cli/v2"
)

// 全局变量
var (
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
	globalFlagForFront *cli.BoolFlag = &cli.BoolFlag{
		Name:    "front",
		Aliases: []string{"f"},
		Value:   false,
		Usage:   "run server on front",
	}
	globalFlagForRedriect *cli.BoolFlag = &cli.BoolFlag{
		Name:    "redriect",
		Aliases: []string{"r"},
		Value:   false,
		Usage:   "run server on front and log into /var/log/${APP}.log, comflict with front",
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
			globalFlagForFront,
			globalFlagForRedriect,
		},
		Action: func(c *cli.Context) error {
			if c.Bool("front") {
				// 前台运行，输出结果到终端
				service.PutStreamServerServiceSetup(ctx, cancel, c.App.Name, port)
				return nil
			}
			if c.Bool("redriect") {
				// 前台运行输出结果到日志文件
				// 日志重定向
				logFile := filepath.Join("/var/log", c.App.Name+".log")
				f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
				if err != nil {
					return nil
				}
				defer f.Close()
				log.SetOutput(f)
				log.Infof("%sd start \n", c.App.Name)
				time.Sleep(2 * time.Second)
				//运行
				service.PutStreamServerServiceSetup(ctx, cancel, c.App.Name, port)
				return nil
			}
			// 后台运行
			// 重新设置命令行参数，删除-r/--redriect
			args := make([]string, 0)
			for _, arg := range os.Args {
				if arg == "-r" || arg == "--redriect" {
					continue
				}
				args = append(args, arg)
			}
			cmdName := args[0]
			cmdArgs := []string{"-r"}
			if len(args) > 1 {
				cmdArgs = append(cmdArgs, args[1:]...)
			}
			cmd := exec.Command(cmdName, cmdArgs...)
			return cmd.Start()
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
	// front run for log
	return nil
}
