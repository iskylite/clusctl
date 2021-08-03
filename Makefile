APP="myclush"

proto:
	protoc -I ./pb/ --go_out=plugins=grpc:. ./pb/stream.proto

build:
	go build -o cmd/${APP}-x64

arm:
	GOOS=linux GOARCH=arm64 go build -o cmd/${APP}-arm64

clean:
	rm -rf cmd/${APP}-x64
	rm -rf cmd/${APP}-arm64

