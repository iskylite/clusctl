// 命令行配置
package main

import (
	"context"
	"os"
	"runtime"
	"sort"

	"github.com/urfave/cli/v2"
)

// 全局基本配置
var (
	name         string = "yhclush"
	version      string = "v1.4.1"
	author       string = "iskylite"
	email        string = "yantao0905@outlook.com"
	descriptions string = "cluster manager tools by grpc service"
)

// 全局默认
// var action func(c *cli.Context) error = func(c *cli.Context) error {
// 	return nil
// }

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
		Value:   2,
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
			PingClientServiceSetup(ctx, nodes, port, c.Int("workers"), c.Int("timeout"))
			return nil
		},
	}
	// 开启服务端子命令 serve
	// 子命令 serve 配置
	serveCommandConfig *cli.Command = &cli.Command{
		Name:    "serve",
		Aliases: []string{"S"},
		Usage:   "start server",
		Action: func(c *cli.Context) error {
			putStreamServerServiceSetup(ctx, cancel, c.App.Name, port)
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
		Value:   "512k",
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
		},
		Action: func(c *cli.Context) error {
			putStreamClientServiceSetup(ctx, cancel, c.String("file"), c.String("dest"), nodes, c.String("size"), port, c.Int("width"))
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
	// 子命令 exec 配置
	execCommandConfig *cli.Command = &cli.Command{
		Name:    "execute",
		Aliases: []string{"exec", "e"},
		Usage:   "execute linux shell command on remote host",
		Flags: []cli.Flag{
			execFlagForCmd,
			execFlagForList,
			execFlagForWidth,
		},
		Action: func(c *cli.Context) error {
			RunCmdClientServiceSetup(ctx, cancel, c.String("cmd"), nodes, c.Int("width"), port, c.Bool("list"))
			return nil
		},
	}
	// 节点带宽测试子命令 oo_bw
	// 子命令 oo_bw 参数配置
	oobwFlagForNodes *cli.StringFlag = &cli.StringFlag{
		Name:     "nodes",
		Aliases:  []string{"N"},
		Usage:    "remote agent host for oo_bw test",
		Required: true,
	}
	oobwFlagForBlockSize *cli.StringFlag = &cli.StringFlag{
		Name:    "blockSize",
		Aliases: []string{"b"},
		Usage:   "block size (16x) for oo_bw test",
		Value:   "0x100000",
	}
	// 子命令 oo_bw 配置
	oobwCommandConfig *cli.Command = &cli.Command{
		Name:    "oo_bw",
		Aliases: []string{"o"},
		Usage:   "normal oo_bw test for point to point",
		Flags: []cli.Flag{
			oobwFlagForBlockSize,
			oobwFlagForNodes,
		},
		Action: func(c *cli.Context) error {
			OoBwServiceSetup(ctx, nodes, c.String("nodes"), c.String("blockSize"), int64(2), false, port, 15, 10)
			return nil
		},
	}
	// 节点带宽循环测试子命令 loop_bw
	loopbwFlagForNodes *cli.StringFlag = &cli.StringFlag{
		Name:     "nodes",
		Aliases:  []string{"N"},
		Usage:    "remote agent host for loop_bw test",
		Required: true,
	}
	loopbwFlagForBlockSize *cli.StringFlag = &cli.StringFlag{
		Name:    "blockSize",
		Aliases: []string{"b"},
		Usage:   "block size (16x) for loop_bw test",
		Value:   "0x100000",
	}
	loopbwFlagForAfterTimes *cli.Int64Flag = &cli.Int64Flag{
		Name:    "after",
		Aliases: []string{"a"},
		Usage:   "run loop_bw after some seconds at the same time",
		Value:   2,
	}
	loopbwFlagForLength *cli.IntFlag = &cli.IntFlag{
		Name:    "len",
		Aliases: []string{"l"},
		Usage:   "loop_bw offset length (8 >> `LEN`)",
		Value:   15,
	}
	loopbwFlagForCounts *cli.IntFlag = &cli.IntFlag{
		Name:    "count",
		Aliases: []string{"c"},
		Usage:   "loop_bw test counts",
		Value:   10,
	}
	// 子命令 loop_bw 配置
	loopbwCommandConfig *cli.Command = &cli.Command{
		Name:    "loop_bw",
		Aliases: []string{"O"},
		Usage:   "loop_bw test",
		Flags: []cli.Flag{
			loopbwFlagForAfterTimes,
			loopbwFlagForBlockSize,
			loopbwFlagForCounts,
			loopbwFlagForLength,
			loopbwFlagForNodes,
		},
		Action: func(c *cli.Context) error {
			OoBwServiceSetup(ctx, nodes, c.String("nodes"), c.String("blockSize"), c.Int64("after"), true, port, c.Int("len"), c.Int("count"))
			return nil
		},
	}
)

func run(ctx context.Context, cancel context.CancelFunc) error {
	app := &cli.App{
		// 基本信息
		Name:     name,
		HelpName: name,
		Version:  version,
		// Description: descriptions,
		Usage: descriptions,
		// 子命令执行前的设置
		Before: func(c *cli.Context) error {
			setLogLevel(c.Bool("debug"))
			return nil
		},
		Authors: []*cli.Author{
			{
				Name:  author,
				Email: email,
			},
		},
		// 全局选项参数配置
		Flags: []cli.Flag{
			globalFlagForNodes,
			globalFlagForDebug,
			globalFlagForPort,
		},
		// 子命令配置
		Commands: []*cli.Command{
			pingCommandConfig,
			serveCommandConfig,
			rcopyCommandConfig,
			execCommandConfig,
			oobwCommandConfig,
			loopbwCommandConfig,
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	return err
}
