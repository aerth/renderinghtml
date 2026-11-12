
PREFIX ?= .
LRH := -L../../librenderinghtml/bin -I../../librenderinghtml/bin -lrenderinghtml
CFLAGS := $(shell pkg-config --cflags --libs gtk4 webkitgtk-6.0) ${LRH}

${PREFIX}/bin/markdown-viewer-gtk: src/*.cpp
	gcc -o $@ $(CFLAGS) $^ 
	sha256sum $@
clean:
	rm -rf bin
