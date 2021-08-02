APP="myclush"

proto:
	protoc -I ./pb/ --go_out=plugins=grpc:. ./pb/stream.proto

build:
	go build -o ${APP}

arm:
	GOOS=linux GOARCH=arm64 go build -o ${APP}

install:
	cp ${APP} /usr/local/sbin
	cp systemd/${APP}.service /usr/lib/systemd/system/${APP}.service

service:
	cp -r systemd/${APP}.service /usr/lib/systemd/system/${APP}.service

clean:
	rm -rf ${APP}
	rm -rf /usr/local/sbin/${APP}
	rm -rf /usr/lib/systemd/system/${APP}.service

pkg-arm64:
	tar -jcf ${APP}_aarch64.tar.bz2 /usr/local/sbin/${APP} /usr/lib/systemd/system/${APP}.service

pkg:
	tar -jcf ${APP}_x64.tar.bz2 /usr/local/sbin/${APP} /usr/lib/systemd/system/${APP}.service
