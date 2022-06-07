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

// 全局选项参数配置
var (
	globalFlagForNodes *cli.StringFlag = &cli.StringFlag{
		Name:    "nodes",
		Aliases: []string{"n"},
		Usage:   "`NODES` where to run the command",
	}
	globalFlagForHostFile *cli.StringFlag = &cli.StringFlag{
		Name:    "hostfile",
		Aliases: []string{"H"},
		Usage:   "path to `FILE` containing a list of target hosts",
	}
	globalFlagForDebug *cli.BoolFlag = &cli.BoolFlag{
		Name:    "debug",
		Aliases: []string{"D"},
		Value:   false,
		Usage:   "set log level debug",
	}
	globalFlagForPort *cli.IntFlag = &cli.IntFlag{
		Name:    "port",
		Aliases: []string{"P"},
		Value:   1995,
		Usage:   "grpc service `PORT`",
	}
	globalFlagForColor *cli.BoolFlag = &cli.BoolFlag{
		Name:    "disablecolor",
		Aliases: []string{"dc"},
		Value:   false,
		Usage:   "disable log color print",
	}
	globalFlagForWidth *cli.IntFlag = &cli.IntFlag{
		Name:    "width",
		Aliases: []string{"w"},
		Value:   2,
		Usage:   "transport tree `WIDTH` for multi workers",
	}
	// 客户端健康检查子命令 ping
	// 子命令 ping 的参数配置
	// 子命令 ping 配置
	globalFlagForPing *cli.BoolFlag = &cli.BoolFlag{
		Name:    "ping",
		Aliases: []string{"p"},
		Value:   false,
		Usage:   "[action] check all agent status",
	}
	pingFlagForTimeout *cli.IntFlag = &cli.IntFlag{
		Name:    "timeout",
		Aliases: []string{"t"},
		Value:   1,
		Usage:   "`TIMEOUT` for ping",
	}

	pingFlagForFanout *cli.IntFlag = &cli.IntFlag{
		Name:    "fanout",
		Aliases: []string{"f"},
		Value:   runtime.NumCPU(),
		Usage:   "use a specified `FANOUT` for ping",
	}
	globalFlagForOutput *cli.StringFlag = &cli.StringFlag{
		Name:    "output",
		Aliases: []string{"o"},
		Usage:   "dump rcopy or command output into log `FILE`",
	}
	// 远程拷贝文件子命令 rcopy
	// 子命令 rcopy 的参数配置
	globalFlagForRCopy *cli.StringFlag = &cli.StringFlag{
		Name:    "rcopy",
		Aliases: []string{"r"},
		Usage:   "[action] local `FILE` path",
	}
	rcopyFlagForDestdir *cli.StringFlag = &cli.StringFlag{
		Name:    "dest",
		Aliases: []string{"d"},
		Usage:   "dest `DIR` on remote host",
		Value:   global.PWD,
	}
	rcopyFlagForBufferSize *cli.StringFlag = &cli.StringFlag{
		Name:    "size",
		Aliases: []string{"s"},
		Usage:   "payload `SIZE` (eg: 51200, 512k, 1m)",
		Value:   "2M",
	}
	// 远程执行子命令 exec
	// 子命令exec 参数配置
	globalFlagForCommand *cli.StringFlag = &cli.StringFlag{
		Name:    "command",
		Aliases: []string{"c"},
		Usage:   "[action] linux shell `COMMAND` to run",
	}
	commandFlagForList *cli.BoolFlag = &cli.BoolFlag{
		Name:    "list",
		Aliases: []string{"l"},
		Usage:   "sort command output by node list",
		Value:   false,
	}
	commandFlagForBackground *cli.BoolFlag = &cli.BoolFlag{
		Name:    "background",
		Aliases: []string{"b"},
		Usage:   "run cmd in background",
		Value:   false,
	}
)

func run(ctx context.Context, cancel context.CancelFunc) error {
	app := &cli.App{
		// 基本信息
		// Name:     name,
		// HelpName: name,
		Version: global.VERSION,
		// Description: descriptions,
		Usage: global.DESC,
		// 子命令执行前的设置
		Before: Before,
		Authors: []*cli.Author{
			{
				Name:  global.AUTHOR,
				Email: global.EMAIL,
			},
		},
		// 全局选项参数配置
		Flags: []cli.Flag{
			globalFlagForNodes,
			globalFlagForHostFile,
			globalFlagForDebug,
			globalFlagForPort,
			globalFlagForColor,
			globalFlagForPing,
			globalFlagForCommand,
			globalFlagForRCopy,
			globalFlagForWidth,
			globalFlagForOutput,
			// ping
			pingFlagForTimeout,
			pingFlagForFanout,
			// rcopy
			rcopyFlagForBufferSize,
			rcopyFlagForDestdir,
			// command
			commandFlagForBackground,
			commandFlagForList,
		},
		// 子命令配置
		Action: func(c *cli.Context) error {
			nodes, err := getNodes(c)
			if err != nil {
				return err
			}
			port := c.Int("port")
			if c.Bool("ping") {
				service.PingClientServiceSetup(ctx, nodes, port, c.Int("fanout"), c.Int("timeout"))
			} else if c.String("rcopy") != "" {
				logfile := c.String("output")
				if logfile != "" {
					f, err := log.SetOutputFile(logfile)
					if err != nil {
						return err
					}
					defer f.Close()
					log.Infof("start: %v\n", os.Args[:])
				}
				service.PutStreamClientServiceSetup(ctx, cancel, c.String("rcopy"), c.String("dest"), nodes,
					c.String("size"), port, c.Int("width"))
			} else if c.String("command") != "" {
				logfile := c.String("output")
				if logfile != "" {
					f, err := log.SetOutputFile(logfile)
					if err != nil {
						return err
					}
					defer f.Close()
					log.Infof("start: %v\n", os.Args[:])
				}
				service.RunCmdClientServiceSetup(ctx, cancel, c.String("command"), nodes, c.Int("width"),
					port, c.Bool("list"), c.Bool("background"))
			}
			return nil
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))

	err := app.Run(os.Args)
	return err
}

// Before pre handler
func Before(c *cli.Context) error {
	// log debug
	log.SetLogLevel(c.Bool("debug"))
	if c.Bool("disablecolor") {
		log.DisableColor()
	}
	if err := handleExclusiveArgs(c); err != nil {
		return err
	}
	// root privileges
	uid, gid, err := utils.UserInfo()
	if err != nil {
		return err
	}
	if uid != "0" && gid != "0" {
		return errors.New("Usage: permission denied, need root privileges")
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

// handleExclusiveArgs 互斥参数检查
func handleExclusiveArgs(c *cli.Context) error {
	var count int
	if c.Bool("ping") {
		count++
	}
	if c.String("rcopy") != "" {
		count++
	}
	if c.String("command") != "" {
		count++
	}

	if count > 1 {
		// 不能同时指定ping、command和rcopy
		return errors.New("Usage: only one action need")
	}
	if c.String("nodes") != "" && c.String("hostfile") != "" {
		// 不可同时指定nodes和hostfile
		return errors.New("Usage: one of --nodes/-n and --hostfile/-H option need")
	}
	return nil
}

// getNodes 获取节点列表
func getNodes(c *cli.Context) (string, error) {
	nodes := c.String("nodes")
	if nodes == "" {
		hostFile := c.String("hostfile")
		if hostFile == "" {
			return nodes, errors.New("Usage: one of --nodes/-n and --hostfile/-H option need")
		}
		nodeList, err := utils.ExpNodesFromFile(hostFile)
		if err != nil {
			return nodes, err
		}
		nodes = utils.Merge(nodeList...)
	}
	return nodes, nil
}
