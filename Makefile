APP = myclush

pb:
	protoc -I ./pb/ --go_out=plugins=grpc:. ./pb/stream.proto

tls:
	openssl req -new -newkey rsa:2048 -days 365000 -nodes -x509 -keyout ./conf/cert.key -out ./conf/cert.pem -extensions v3_req -config ./conf/openssl.cnf

build:
	go build -o ${APP} cmd/client/*.go
	go build -o ${APP}d cmd/server/*.go
arm:
	GOOS=linux GOARCH=arm64 go build -o ${APP}-arm64 cmd/client/*.go
	GOOS=linux GOARCH=arm64 go build -o ${APP}d-arm64 cmd/server/*.go

clean:
	rm -rf ${APP}-x64 ${APP}d-x64
	rm -rf ${APP}-arm64 ${APP}d-arm64

