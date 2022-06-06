PHONY: help

APP = myclush

protoc:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=require_unimplemented_servers=false:. --go-grpc_opt=paths=source_relative pb/stream.proto

tls-centos:
	openssl req -new -newkey rsa:2048 -days 36500 -nodes -x509 -keyout ./conf/cert.key -out ./conf/cert.pem -extensions v3_req -config ./conf/openssl.cnf.centos

tls-ubuntu:
	openssl req -new -newkey rsa:2048 -days 36500 -nodes -x509 -keyout ./conf/cert.key -out ./conf/cert.pem -extensions v3_req -config ./conf/openssl.cnf.ubuntu

build:
	go build -o bin/${APP} cmd/client/*.go
	go build -o bin/${APP}d cmd/server/*.go

arm:
	GOOS=linux GOARCH=arm64 go build -o bin/${APP}-arm64 cmd/client/*.go
	GOOS=linux GOARCH=arm64 go build -o bin/${APP}d-arm64 cmd/server/*.go

clean:
	rm -rf bin/*

help:
	@echo "usage: make [protoc|tls-centos|tls-ubuntu|build|arm|clean]"

