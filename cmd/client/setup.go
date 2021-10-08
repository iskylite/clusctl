// 命令行配置
package main

import (
	"context"
	"errors"
	"myclush/service"
	"myclush/utils"
	"os"
	"runtime"
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
	// 是否有颜色输出
	color bool
)

// 全局选项参数配置
var (
	globalFlagForNodes *cli.StringFlag = &cli.StringFlag{
		Name:        "nodes",
		Aliases:     []string{"n"},
		Usage:       "app agent nodes list",
		Destination: &nodes,
	}
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
	globalFlagForColor *cli.BoolFlag = &cli.BoolFlag{
		Name:        "disablecolor",
		Aliases:     []string{"dc"},
		Value:       false,
		Usage:       "disable log color print",
		Destination: &color,
	}
)

// 子命令设置
var (
	// 客户端健康检查子命令 ping
	// 子命令 ping 的参数配置
	pingFlagForWorkers *cli.IntFlag = &cli.IntFlag{
		Name:    "workers",
		Aliases: []string{"w"},
		Value:   runtime.NumCPU(),
		Usage:   "ping goroutine counts at the same time",
	}
	pingFlagForTimeout *cli.IntFlag = &cli.IntFlag{
		Name:    "timeout",
		Aliases: []string{"t"},
		Value:   1,
		Usage:   "timeout for check agent status",
	}
	// 子命令 ping 配置
	pingCommandConfig *cli.Command = &cli.Command{
		Name:    "ping",
		Aliases: []string{"P"},
		Usage:   "check all agent status",
		Flags: []cli.Flag{
			pingFlagForWorkers,
			pingFlagForTimeout,
		},
		Action: func(c *cli.Context) error {
			service.PingClientServiceSetup(ctx, nodes, port, c.Int("workers"), c.Int("timeout"))
			return nil
		},
	}
	// 远程拷贝文件子命令 rcopy
	// 子命令 rcopy 的参数配置
	rcopyFlagForFile *cli.StringFlag = &cli.StringFlag{
		Name:     "file",
		Aliases:  []string{"f"},
		Usage:    "local `FILE` path",
		Required: true,
	}
	rcopyFlagForDestdir *cli.StringFlag = &cli.StringFlag{
		Name:    "dest",
		Aliases: []string{"d"},
		Usage:   "dest `DIR` on remote host",
		Value:   "/tmp",
	}
	rcopyFlagForWidth *cli.IntFlag = &cli.IntFlag{
		Name:    "width",
		Aliases: []string{"w"},
		Usage:   "B+ tree width for transmission data",
		Value:   50,
	}
	rcopyFlagForBufferSize *cli.StringFlag = &cli.StringFlag{
		Name:    "size",
		Aliases: []string{"b", "s"},
		Usage:   "payload size (eg: 51200, 512k, 1m) in rpc package",
		Value:   "2M",
	}
	rcopyFlagForOutput *cli.StringFlag = &cli.StringFlag{
		Name:    "output",
		Aliases: []string{"o"},
		Usage:   "dump output into log file",
	}
	// 子命令 rcopy 配置
	rcopyCommandConfig *cli.Command = &cli.Command{
		Name:    "rcopy",
		Aliases: []string{"rc", "r"},
		Usage:   "copy local file to remote host by grpc service",
		Flags: []cli.Flag{
			rcopyFlagForFile,
			rcopyFlagForBufferSize,
			rcopyFlagForDestdir,
			rcopyFlagForWidth,
			rcopyFlagForOutput,
		},
		Action: func(c *cli.Context) error {
			logfile := c.String("output")
			if logfile != "" {
				f, err := log.SetOutputFile(logfile)
				if err != nil {
					return err
				}
				defer f.Close()
				log.Infof("start: %v\n", os.Args[:])
			}
			service.PutStreamClientServiceSetup(ctx, cancel, c.String("file"), c.String("dest"), nodes, c.String("size"), port, c.Int("width"))
			return nil
		},
	}
	// 远程执行子命令 exec
	// 子命令exec 参数配置
	execFlagForCmd *cli.StringFlag = &cli.StringFlag{
		Name:     "cmd",
		Aliases:  []string{"c"},
		Required: true,
		Usage:    "linux shell command to run",
	}
	execFlagForWidth *cli.IntFlag = &cli.IntFlag{
		Name:    "width",
		Aliases: []string{"w"},
		Usage:   "B+ tree width for executing command",
		Value:   50,
	}
	execFlagForList *cli.BoolFlag = &cli.BoolFlag{
		Name:    "list",
		Aliases: []string{"l"},
		Usage:   "sort command output by node list",
		Value:   false,
	}
	execFlagForBackground *cli.BoolFlag = &cli.BoolFlag{
		Name:    "background",
		Aliases: []string{"b"},
		Usage:   "run cmd in background",
		Value:   false,
	}
	execFlagForOutput *cli.StringFlag = &cli.StringFlag{
		Name:    "output",
		Aliases: []string{"o"},
		Usage:   "dump output into log file",
	}
	// 子命令 exec 配置
	execCommandConfig *cli.Command = &cli.Command{
		Name:    "execute",
		Aliases: []string{"exec", "e"},
		Usage:   "execute linux shell command on remote host",
		Flags: []cli.Flag{
			execFlagForCmd,
			execFlagForList,
			execFlagForWidth,
			execFlagForBackground,
			execFlagForOutput,
		},
		Action: func(c *cli.Context) error {
			logfile := c.String("output")
			if logfile != "" {
				f, err := log.SetOutputFile(logfile)
				if err != nil {
					return err
				}
				defer f.Close()
				log.Infof("start: %v\n", os.Args[:])
			}
			service.RunCmdClientServiceSetup(ctx, cancel, c.String("cmd"), nodes, c.Int("width"), port, c.Bool("list"), c.Bool("background"))
			return nil
		},
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
			globalFlagForNodes,
			globalFlagForDebug,
			globalFlagForPort,
			globalFlagForColor,
		},
		// 子命令配置
		Commands: []*cli.Command{
			pingCommandConfig,
			rcopyCommandConfig,
			execCommandConfig,
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	return err
}

func Before(c *cli.Context) error {
	// log debug
	log.SetLogLevel(c.Bool("debug"))
	if c.Bool("disablecolor") {
		log.DisableColor()
	}
	if c.String("nodes") == "" {
		return errors.New("flag \"--nodes\" or \"-r\" not provide")
	}
	// root privileges
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
	// authority
	authority, err := service.SetAuthorityMetadata()
	if err != nil {
		return err
	}
	global.Authority = authority
	return nil
}
