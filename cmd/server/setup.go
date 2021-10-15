// 命令行配置
package main

import (
	"context"
	"errors"
	"fmt"
	"myclush/service"
	"myclush/utils"
	"net/http"
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
	// pprof
	pprof int
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
	globalFlagForMunalGC *cli.BoolFlag = &cli.BoolFlag{
		Name:        "munalgc",
		Aliases:     []string{"gc"},
		Value:       false,
		Usage:       "munal-gc",
		Destination: &global.MunalGC,
	}
	globalFlagForPprof *cli.IntFlag = &cli.IntFlag{
		Name:        "pprof",
		Aliases:     []string{"pf"},
		Usage:       "pprof web ui",
		Destination: &pprof,
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
			globalFlagForMunalGC,
			globalFlagForPprof,
		},
		Action: func(c *cli.Context) error {
			if c.Bool("front") {
				// 前台运行，输出结果到终端
				service.PutStreamServerServiceSetup(ctx, cancel, c.App.Name, port)
				return nil
			}
			if _, ok := os.LookupEnv("MYCLUSH_DAEMON"); ok {
				// app运行在子进程中
				// 日志重定向
				logFile := filepath.Join("/var/log", c.App.Name+".log")
				f, err := log.SetOutputFile(logFile)
				if err != nil {
					return err
				}
				defer f.Close()
				log.Infof("%s start \n", c.App.Name)
				time.Sleep(2 * time.Second)
				//运行
				service.PutStreamServerServiceSetup(ctx, cancel, c.App.Name, port)
			} else {
				// 后台运行
				env := os.Environ()
				cmd := exec.Command(os.Args[0], os.Args[1:]...)
				if _, ok := os.LookupEnv("HOME"); !ok {
					env = append(env, "HOME=/root")
				}
				env = append(env, "MYCLUSH_DAEMON=on")
				cmd.Env = env
				return cmd.Start()
			}
			return nil
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	return err
}

func Before(c *cli.Context) error {
	if c.Bool("debug") {
		log.SetLevel(log.DEBUG)
	}
	if c.Bool("munalgc") {
		log.Info("enable put stream munal-gc")
	} else {
		log.Info("disable put stream munal-gc")
	}
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

	// pprof
	if c.Int("pprof") != 0 {
		go func() {
			if err := http.ListenAndServe(fmt.Sprintf(":%d", c.Int("pprof")), nil); err != nil {
				log.Error("funcRetErr=http.ListenAndServe||err=%s", err.Error())
			}
		}()
		log.Infof("enable pprof on %d\n", c.Int("pprof"))
	} else {
		log.Info("disable pprof")
	}
	return nil
}
