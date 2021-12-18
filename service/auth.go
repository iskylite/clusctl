package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"myclush/logger"
	"myclush/utils"
	"os"
	"os/user"
	"path/filepath"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// 一元拦截器
func unaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// 查询服务端地址
		if ip, err := clientIP(ctx); err == nil {
			logger.Infof("unary client: [%s]\n", ip)
		} else {
			logger.Error(err)
		}
		// auth
		if err := authBySSHKeys(ctx); err != nil {
			logger.Errorf("unary path: [%s] validate failed: %v\n", info.FullMethod, err)
			return nil, err
		}
		logger.Infof("unary path: [%s] validate passed\n", info.FullMethod)
		return handler(ctx, req)
	}
}

// stream拦截器
func streamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// 查询服务端地址
		if ip, err := clientIP(ss.Context()); err == nil {
			logger.Infof("unary client: [%s]\n", ip)
		} else {
			logger.Error(err)
		}
		// auth
		if err := authBySSHKeys(ss.Context()); err != nil {
			// 由于直接返回的错误在repliesChannel中是没有节点信息的，所以不知到是哪个节点出错了，故在此处直接返回错误信息
			// 并且对于stream.Recv接受普通的错误不做处理
			logger.Errorf("stream path: [%s] validate failed: %v\n", info.FullMethod, err)
			ss.SendMsg(newReply(false, utils.GrpcErrorMsg(err), utils.Hostname()))
			return err
		}
		logger.Infof("stream path: [%s] validate passed\n", info.FullMethod)
		return handler(srv, ss)
	}
}

// 依据ssh公钥认证
func authBySSHKeys(ctx context.Context) error {
	sshKey, err := getAuthorityByContext(ctx)
	if err != nil {
		return status.Error(codes.Unauthenticated, err.Error())
	}
	if err := CheckLocalSSHAuthorizedKeys(sshKey); err != nil {
		return status.Error(codes.Unauthenticated, err.Error())
	}
	return nil
}

// 获取当前用户的
func LocalUserSSHKey() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return homedir, err
	}
	ssh_id_rsa_pub := filepath.Join(homedir, ".ssh", "id_rsa.pub")
	f, err := os.Open(ssh_id_rsa_pub)
	if err != nil {
		return homedir, err
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		return homedir, err
	}
	return utils.Md5sum(bytes.TrimSpace(content)), nil
}

func CheckLocalSSHAuthorizedKeys(sshKeysMd5Str string) error {
	// sometime UserHomeDir will panic: $HOME not defined
	// homedir := "/root"
	user, err := user.Current()
	if err != nil {
		return err
	}
	homedir := user.HomeDir
	// not use os.UerHomeDir to avoid $HOME not defined when control by systemd
	// homedir, err := os.UserHomeDir()
	// if err != nil {
	// 	return err
	// }
	authorizedKeys := filepath.Join(homedir, ".ssh", "authorized_keys")
	f, err := os.Open(authorizedKeys)
	if err != nil {
		return err
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	allAuthorizedKeys := bytes.Split(content, []byte{'\n'})
	for _, authorizedKey := range allAuthorizedKeys {
		if len(authorizedKey) == 0 {
			continue
		}
		localMd5str := utils.Md5sum(bytes.TrimSpace(authorizedKey))
		if localMd5str == sshKeysMd5Str {
			return nil
		}
	}
	return errors.New("validate ssh_keys failed")
}
func getAuthorityByContext(ctx context.Context) (string, error) {
	var sshKey string
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		err := errors.New("no metadata")
		return sshKey, err
	}
	sshKeys, ok := md["token"]
	if !ok {
		err := errors.New("no ssh token")
		return sshKey, err
	}
	sshKey = sshKeys[0]
	return sshKey, nil
}

// when tls, authority not work
func SetAuthority() (grpc.DialOption, error) {
	sshKeys, err := LocalUserSSHKey()
	if err != nil {
		return grpc.EmptyDialOption{}, err
	}
	return grpc.WithAuthority(sshKeys), nil
}

// when tls, use grpc.WithPerRPCCredentials
type authority struct {
	sshKey string
}

// GetRequestMetadata 获取当前请求认证所需的元数据（metadata）
func (a *authority) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"token": a.sshKey}, nil
}

// RequireTransportSecurity 是否需要基于 TLS 认证进行安全传输
func (a *authority) RequireTransportSecurity() bool {
	return true
}

func SetAuthorityMetadata() (grpc.DialOption, error) {
	sshKey, err := LocalUserSSHKey()
	if err != nil {
		return nil, err
	}
	return grpc.WithPerRPCCredentials(&authority{sshKey: sshKey}), nil
}

func clientIP(ctx context.Context) (string, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", errors.New("no peer information exists")
	}
	return p.Addr.String(), nil
}
