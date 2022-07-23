PHONY: help

APP = clusctl
VERSION = 1.6.2
BUILDPATH = build
ARCH = `uname -m`
RELEASE = 0

protoc:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=require_unimplemented_servers=false:. --go-grpc_opt=paths=source_relative pb/stream.proto

tls-centos:
	rm -rf conf/cert.key conf/cert.pem
	openssl req -new -newkey rsa:2048 -days 36500 -nodes -x509 -keyout ./conf/cert.key -out ./conf/cert.pem -extensions v3_req -config ./conf/openssl.cnf.centos -subj "/C=CN/ST=HN/L=ZZ/O=Sugon/OU=HPC/CN=myclush.com"

tls-ubuntu:
	rm -rf conf/cert.key conf/cert.pem
	openssl req -new -newkey rsa:2048 -days 36500 -nodes -x509 -keyout ./conf/cert.key -out ./conf/cert.pem -extensions v3_req -config ./conf/openssl.cnf.ubuntu -subj "/C=CN/ST=HN/L=ZZ/O=Sugon/OU=HPC/CN=myclush.com"

x64:
	go build -o bin/${APP} cmd/client/*.go
	go build -o bin/${APP}d cmd/server/*.go

rpm:
	rm -rf ${BUILDPATH}
	ls conf/{cert.key,cert.pem} || make tls-centos
	@echo "abort make tls-centos"
	sed -i "s/VERSION\ string\ =\ .*/VERSION\ string\ =\ \"v${VERSION}\"/g" global/var.go
	make build
	mkdir -p ${BUILDPATH}/{BUILD,BUILDROOT,RPMS,SOURCES,SPECS,SRPMS}
	mkdir -p ${BUILDPATH}/BUILD/${APP}-${VERSION}
	mkdir -p ${BUILDPATH}/BUILD/${APP}-${VERSION}/usr/sbin
	mkdir -p ${BUILDPATH}/BUILD/${APP}-${VERSION}/etc/systemd/system
	mkdir -p ${BUILDPATH}/BUILD/${APP}-${VERSION}/var/lib/${APP}d
	sed -i 's/clusctld/${APP}d/g' conf/${APP}d.service
	/bin/cp -af bin/* ${BUILDPATH}/BUILD/${APP}-${VERSION}/usr/sbin
	/bin/cp -af conf/${APP}d.service ${BUILDPATH}/BUILD/${APP}-${VERSION}/etc/systemd/system
	/bin/cp -af conf/{cert.pem,cert.key} ${BUILDPATH}/BUILD/${APP}-${VERSION}/var/lib/${APP}d
	sed -i "s/version\ .*/version ${VERSION}/g" conf/clusctl.spec 
	sed -i "s/release\ .*/release ${RELEASE}/g" conf/clusctl.spec 
	rpmbuild -bb conf/${APP}.spec --define="_topdir `pwd`/${BUILDPATH}"
	mv ${BUILDPATH}/RPMS/`uname -m`/${APP}-${VERSION}-${RELEASE}.${ARCH}.rpm ./

clean:
	rm -rf bin/*

help:
	@echo "usage: make [protoc|tls-centos|tls-ubuntu|build|rpm|clean]"

