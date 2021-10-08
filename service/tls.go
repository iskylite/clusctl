package service

import (
	"myclush/global"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// CLientTLS
func GenClientTransportCredentials() (grpc.DialOption, error) {
	creds, err := credentials.NewClientTLSFromFile(global.CertPemPath, "myclush.com")
	if err != nil {
		return grpc.EmptyDialOption{}, err
	}
	return grpc.WithTransportCredentials(creds), nil
}

// server TLS
func GenServerTransportCredentials() (grpc.ServerOption, error) {
	creds, err := credentials.NewServerTLSFromFile(global.CertPemPath, global.CertKeyPath)
	if err != nil {
		return grpc.EmptyServerOption{}, err
	}
	return grpc.Creds(creds), nil
}
