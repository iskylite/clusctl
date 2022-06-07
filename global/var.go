package global

import (
	"fmt"
	"os"

	"google.golang.org/grpc"
)

// 全局基本配置
const (
	APP     string = "clusctl"
	VERSION string = "v1.6.0"
	AUTHOR  string = "iskylite"
	EMAIL   string = "yantao0905@outlook.com"
	DESC    string = "HPC Cluster Manager Tools"
	SUCCESS string = "Success"
)

var (
	// Authority grpc Authority DialOption
	Authority grpc.DialOption
	// ClientTransportCredentials grpc client cred DialOption
	ClientTransportCredentials grpc.DialOption
	// ServerTransportCredentials grpc server cred DialOption
	ServerTransportCredentials grpc.ServerOption
	// MunalGC munally run gc
	MunalGC bool // 是否手动gc
	// PWD local path
	PWD string
)

var (
	// CertKeyPath cert key path
	CertKeyPath string = fmt.Sprintf("/var/lib/%sd/cert.key", APP)
	// CertPemPath cert pem path
	CertPemPath string = fmt.Sprintf("/var/lib/%sd/cert.pem", APP)
)

func init() {
	var err error
	PWD, err = os.Getwd()
	if err != nil {
		panic(err)
	}
}
