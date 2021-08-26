// 命令行配置
package global

import "google.golang.org/grpc"

// 全局基本配置
const (
	Version      string = "v1.5.0"
	Author       string = "iskylite"
	Email        string = "yantao0905@outlook.com"
	Descriptions string = "cluster manager tools by grpc service"
	Success      string = "Success"
)

var (
	Authority grpc.DialOption
)
