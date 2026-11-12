PREFIX ?= /usr
build:
	go build -o ./bin/ -buildmode=c-shared .
	mv bin/librenderinghtml bin/librenderinghtml.so
install:
	install bin/librenderinghtml.so ${PREFIX}/lib/librenderinghtml.so
	install bin/librenderinghtml.h ${PREFIX}/include/librenderinghtml.h
clean:
	rm -rf bin
