APP = myclush

proto:
	protoc -I ./pb/ --go_out=plugins=grpc:. ./pb/stream.proto

build:
	cd cmd/server
	go build . 
	cd -
	cd cmd/server
	go build . 
	cd -
arm:
	cd cmd/server
	GOOS=linux GOARCH=arm64 go build -o ${APP}d-arm64
	cd -
	cd cmd/server
	GOOS=linux GOARCH=arm64 go build -o ${APP}-arm64
	cd -

clean:
	rm -rf cmd/${APP}-x64
	rm -rf cmd/${APP}-arm64

