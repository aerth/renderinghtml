flags ?= 
#-gcflags=-m
build:
	go build -v -trimpath -ldflags "-w -s" ${flags} -o bin/ ./cmd/...
run: build
	bin/tester -addr 127.0.0.1:8081 -d . -serve-readme
library: build
	${MAKE} -C librenderinghtml build 
bin/markdown-viewer-gtk: build example/*/*
	make -C librenderinghtml
	ls -l librenderinghtml/
	make -C example/markdown-viewer-gtk PREFIX=$(PWD)
	@echo Be sure to run: sudo make -C librenderinghtml install
all: build library bin/markdown-viewer-gtk
clean:
	rm -rf bin
	${MAKE} -C librenderinghtml clean
	${MAKE} -C example/markdown-viewer-gtk clean
	
