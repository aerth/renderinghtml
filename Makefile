flags ?= 
#-gcflags=-m
build:
	go build -v -trimpath -ldflags "-w -s" ${flags} -o bin/ ./cmd/...
run: build
	bin/tester -addr 127.0.0.1:8081 -d . -serve-readme
