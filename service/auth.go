package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"myclush/logger"
	"myclush/utils"
	"os"
	"path/filepath"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// 一元拦截器
func unaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if err := authBySSHKeys(ctx); err != nil {
			logger.Errorf("unary [%s] validate failed: %v\n", info.FullMethod, err)
			return nil, err
		}
		logger.Debugf("unary [%s] validate pass\n", info.FullMethod)
		return handler(ctx, req)
	}
}

// stream拦截器
func streamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := authBySSHKeys(ss.Context()); err != nil {
			// 由于直接返回的错误在repliesChannel中是没有节点信息的，所以不知到是哪个节点出错了，故在此处直接返回错误信息
			// 并且对于stream.Recv接受普通的错误不做处理
			logger.Errorf("stream [%s] validate failed: %v\n", info.FullMethod, err)
			ss.SendMsg(newReply(false, utils.GrpcErrorMsg(err), utils.Hostname()))
			return err
		}
		logger.Debugf("stream [%s] validate pass\n", info.FullMethod)
		return handler(srv, ss)
	}
}

// 错误定义
var ()

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
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	authorized_keys := filepath.Join(homedir, ".ssh", "authorized_keys")
	f, err := os.Open(authorized_keys)
	if err != nil {
		return err
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	all_authorized_keys := bytes.Split(content, []byte{'\n'})
	for _, authorized_key := range all_authorized_keys {
		if len(authorized_key) == 0 {
			continue
		}
		localMd5str := utils.Md5sum(bytes.TrimSpace(authorized_key))
		if localMd5str == sshKeysMd5Str {
			return nil
		}
	}
	return errors.New("validate ssh_keys failed")
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

func getAuthorityByContext(ctx context.Context) (string, error) {
	var sshKey string
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		err := errors.New("no metadata")
		return sshKey, err
	}
	sshKeys, ok := md[":authority"]
	if !ok {
		err := errors.New("no authority")
		return sshKey, err
	}
	sshKey = sshKeys[0]
	return sshKey, nil
}

func SetAuthority() (grpc.DialOption, error) {
	sshKeys, err := LocalUserSSHKey()
	if err != nil {
		return grpc.EmptyDialOption{}, err
	}
	return grpc.WithAuthority(sshKeys), nil
}
