# myclush

应用于HPC的多节点数据传输和命令执行

## 初始化环境

```shell
go mod tidy
```

安装 protoc，请自行下载，解压后将 protoc 文件目录添加到系统环境变量 PATH 中。

## 配置 TLS

```shell
# 生成证书和私钥
make tls

# 在服务端和客户端都建立key存放目录： /var/lib/myclushd
mkdir -p /var/lib/myclushd

# 拷贝
cp conf/cert.* /var/lib/myclushd
```

## 编译

```shell
# 编译x64
make build

# 编译arm64
make arm
```

## 运行

### 服务端

```shell
myclushd [-d|-f]
```

- 参数说明:

  - -d 调试

  - -f 前台运行，日志输出到屏幕

### 客户端

```shell
NAME:
   myclush - cluster manager tools by grpc service

USAGE:
   myclush [global options] command [command options] [arguments...]

VERSION:
   v1.5.0

AUTHOR:
   iskylite <yantao0905@outlook.com>

COMMANDS:
   execute, exec, e  execute linux shell command on remote host
   ping, P           check all agent status
   rcopy, rc, r      copy local file to remote host by grpc service
   help, h           Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug, -d              set log level debug (default: false)
   --nodes value, -n value  app agent nodes list
   --port value, -p value   grpc service port (default: 1995)
   --help, -h               show help (default: false)
   --version, -v            print the version (default: false)
```

目前可以使用 systemd 托管 myclushd 服务。

```shell
cp myclushd.service /etc/systemd/system
systemctl daemon-reload
systemctl enable --now myclushd
```

#### 示例

##### 检查服务端是否在线

```shell
./myclush -n vn[0-3] ping
PING: 4/4 vn1

--------------------
vn[0-3]  (4)
--------------------

```

##### 拷贝文件

```shell
# no debug
./myclush -n vn[0-3] rc -f ../../software/WSL2-Linux-Kernel-5.4.zip -d /tmp
数据读取: 102/102 EOF
结果汇总: 4/4 EOF

--------------------
vn[0-3]  (4)
--------------------
Success
# debug
./myclush -d -n vn[0-3] rc -f ../../software/WSL2-Linux-Kernel-5.4.zip -d /tmp
```

##### 远程执行命令

```shell
./myclush -n vn[0-3] exec -c date
结果汇总: 4/4 EOF

--------------------
vn[0-3]  (4)
--------------------
Wed Sep  1 10:42:42 CST 2021

# 如果需要远程执行复杂的shell命令，对于$需要转义，且整个命令使用双引号包裹住
./myclush -n vn[0-3] exec -c "df | grep vmhgfs"
结果汇总: 4/4 EOF

--------------------
vn[0-3]  (4)
--------------------
vmhgfs-fuse    498565120 290707404 207857716  59% /app
./myclush -n vn[0-3] exec -c "df | grep vmhgfs | awk '{print $5}'"
结果汇总: 4/4 EOF

--------------------
vn[0-3]  (4)
--------------------
vmhgfs-fuse    498565120 290707472 207857648  59% /app
./myclush -n vn[0-3] exec -c "df | grep vmhgfs | awk '{print \$5}'"
结果汇总: 4/4 EOF

--------------------
vn[0-3]  (4)
--------------------
59%

# 后台运行命令，提交成功则返回Success
./myclush -n vn[0-3] exec -c "df | grep vmhgfs | awk '{print \$5}'" -b
结果汇总: 4/4 EOF

--------------------
vn[0-3]  (4)
--------------------
Success
```

## 调试

- 使用全局参数 **-d** 可以看到详细的执行信息，便于查找问题。
- 服务端使用 **-f** 参数时 myclushd 运行在前端，日志将会打印到屏幕。否则 myclushd 运行在后台，日志输出到 **/var/log/myclushd.log**。
- 客户端输出在屏幕。

## 测试运行

实际运行中测试过 92160 个计算节点使用 myclush 拷贝数据和执行命令，运行无错误。
