VERSION := $(shell cat VERSION)

default:
	install

install:
	go install .

xcompile:
	env GOOS=linux GOARCH=amd64 go build -ldflags "-X main.VERSION=$(VERSION)" -o goviz_linux
	env GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.VERSION=$(VERSION)" -o goviz_mac
