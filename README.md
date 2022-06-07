# clusctl -- 高性能计算集群计算节点批处理运维工具

## 编译

### 初始化环境

```shell
go mod tidy
```

### 编译RPM

```shell
make rpm
```

编译完成后将会在当前目录生成RPM包：***clusctl-\<version>-\<release>.\<arch>.rpm***


## 安装

```shell
rpm -ivh clusctl-<version>-<release>.<arch>.rpm
```

安装完后将会自动开启服务端，并设置开机自启，即clusctld.service。
对于只使用客户端的节点来说请执行以下命令关闭服务端。

```shell
systemctl disable --now clusctld.service
```

## 使用

> 也可以使用 ***-H/--hostfile*** 指定节点列表文件或者IP列表文件。文件中一行是一个节点名或者IP

### 服务端状态检查

```shell
clusctl -n admin0 -p
PING: 1/1 admin0

--------------------
admin0  (1)
--------------------
```

### 命令执行

```shell
clusctl -n admin0 -c date
结果汇总: 1/1 EOF

--------------------
admin0  (1)
--------------------
Tue Jun  7 19:43:23 CST 2022
```

### 文件传输

```shell
clusctl -n admin0 -r clusctl-1.6.0-1.x86_64.rpm -d /root
数据读取: 4/4 EOF
结果汇总: 1/1 EOF

--------------------
admin0  (1)
--------------------
Success
```