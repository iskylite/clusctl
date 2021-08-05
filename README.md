# myclush

类似于 slurm 的多节点数据传输和命令执行（待定）

# 初始化环境

```shell
go mod tidy
```

安装 protoc，请自行下载，解压后将 protoc 文件目录添加到系统环境变量 PATH 中。

# 编译

```shell
# 编译protobuf
$ make proto

# 编译x64
$ make build

# 编译arm64
$ make arm

# 生成service
$ make service
```

service 也可自行配置。模板如下：

```shell
# /usr/lib/systemd/system/myclush.service
[Unit]
Description=myclush server for remote copy file and exec shell command
After=network.target setup.service

[Service]
Type=simple
ExecStart=/usr/local/sbin/myclush -s -D
StandardOutput=file:/var/log/myclush.log
StandardError=file:/var/log/myclush.log
ExecStop=/bin/kill -s TERM $MAINPID

[Install]
WantedBy=multi-user.target
```

# 运行

## 命令行参数

```shell
$ ./myclush
  -D    debug log
  -P    start ping service
  -W int
        ping  workers max number (default 16)
  -b int
        buffersize bytes (default 524288)
  -c string
        start myclush client and copy file to remote server
  -d string
        destPath (default "/tmp")
  -e string
        command string
  -l    sort cmd output by node list
  -n string
        nodes string
  -p string
        grpc server port (default "1995")
  -s    start myclush server service
  -t int
        command execute timeout (default 3)
  -w int
        B tree width (default 2)

```

## 服务端

服务端安装在计算节点。

```shell
# 如果按照上述步骤配置完成service，可以使用systemd进行调度，日志存放地址： /var/log/myclush.log
$ systemctl start myclush

# 或
# 命令行执行
$ myclush -s

# 调试模式
$ myclush -s -d
```

## 客户端

### 数据传输

```shell
# 示例拷贝文件
$ ./myclush -N cn[0-3,5-9,12-71,74-81,84-125,128-243] -c myclush.tar.gz -d /root
```

如果所有节点正常执行完成，那么将显示一下输出信息：

```shell
PASS: cn[0-3,5-9,12-71,74-81,84-125,128-243], SUM: 235
```

测试结果：

235 个计算节点，传输文件大小 6GB。目前服务端缓冲数量暂时不可修改。

- **树宽为 2，传输大小为 512KB，服务端传输缓冲数量为 64，单点接收速度为 160MB/s, 单点发送速度为 320MB/s， 共用时 41s**
- **树宽为 2，传输大小为 512KB，服务端传输缓冲数量为 0，单点接收速度为 60MB/s, 单点发送速度为 120MB/s， 共用时 1m30s**

- **树宽为 2，传输大小为 512KB，服务端传输缓冲数量为 128，单点接收速度为 160MB/s, 单点发送速度为 320MB/s， 共用时 42s**

- **树宽为 16，传输大小为 512KB，服务端传输缓冲数量为 64，单点接收速度为 60MB/s, 单点发送速度为 900MB/s， 共用时 1m52s**

- **树宽为 2，传输大小为 820KB，服务端传输缓冲数量为 64，单点接收速度为 160MB/s, 单点发送速度为 320MB/s， 共用时 42s**

- **树宽为 2，传输大小为 1MB，服务端传输缓冲数量为 64，单点接收速度为 160MB/s, 单点发送速度为 320MB/s， 共用时 41s**

总结：

建议使用默认参数。

### 命令执行

```shell
# 示例执行命令
# 无法执行交互命令
$ ./myclush -N cn[0-3,5-9,12-71,74-81,84-125,128-243] -e date

# 输出结果按照节点列表顺序输出
$ ./myclush -N cn[0-3,5-9,12-71,74-81,84-125,128-243] -e date -l
```

如果所有节点正常执行完成，那么将显示一下输出信息：

```shell
>>> cn76 [PASS]
Fri Jun 25 19:57:46 CST 2021

>>> cn77 [PASS]
Fri Jun 25 19:57:46 CST 2021

>>> cn78 [PASS]
Fri Jun 25 19:57:46 CST 2021

PASS: cn[0-3,5-9,12-71,74-81,84-125,128-243], SUM: 235
```

### 检查节点健康状况

```shell
# 示例执行命令
$ ./myclush -N cn[0-3,5-9,12-71,74-81,84-125,128-243] -P
```

如果所有节点正常执行完成，那么将显示一下输出信息：

```shell
PASS: cn[0-3,5-9,12-71,74-81,84-125,128-243], SUM: 235
```

### FAQ

- GRPC 调试日志

需要设置环境变量

```shell
export GRPC_GO_LOG_SEVERITY_LEVEL="info"
export GRPC_GO_LOG_VERBOSITY_LEVEL=2
```

- 执行命令有正确输出但是仍然显示 failed

部分命令执行过程中有报错信息，故仍会显示 failed，如下为 df 命令的示例：

```shell

>>> cn20 [FAILED]
df: /mnt/glex: Transport endpoint is not connected
Filesystem            1K-blocks     Used   Available Use% Mounted on
ramfs                  65205872  7443824    57762048  12% /
devtmpfs               65074576        0    65074576   0% /dev
tmpfs                  65205872        0    65205872   0% /dev/shm
tmpfs                  13041176      420    13040756   1% /run
tmpfs                      5120        0        5120   0% /run/lock
tmpfs                  65205872        0    65205872   0% /sys/fs/cgroup
tmpfs                  65205872        0    65205872   0% /var/volatile
tmpfs                  13041172        0    13041172   0% /run/user/0

PASS: cn[0-3,5-7,9,12-19,21-29,31-39,41-71,74-81,84-125,128-243], SUM: 231
FAILED: cn[8,20,30,40], SUM: 4
```

- 默认值

默认树宽为 2，即二叉树
默认传输大小为 512KB

## TODO

1、可以重用 client

```shell
grpcOptions := grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024 * 1024 * 3 * len(nodes)))
 conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpcOptions)
 if err != nil {
  return nil, utils.GrpcErrorWrapper(err)
 }
 log.Debugf("Dial Server %s\n", addr)
 client := pb.NewRpcServiceClient(conn)
```
