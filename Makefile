.PHONY: build build-windows build-linux build-darwin clean install

# 默认构建 Linux 版本（主要使用）
build:
	GOOS=linux GOARCH=amd64 go build -o bin/l2h-s ./cmd/l2h-s
	GOOS=linux GOARCH=amd64 go build -o bin/l2h-c ./cmd/l2h-c

build-windows:
	go build -o bin/l2h-s.exe ./cmd/l2h-s
	go build -o bin/l2h-c.exe ./cmd/l2h-c

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/l2h-s ./cmd/l2h-s
	GOOS=linux GOARCH=amd64 go build -o bin/l2h-c ./cmd/l2h-c

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o bin/l2h-s ./cmd/l2h-s
	GOOS=darwin GOARCH=amd64 go build -o bin/l2h-c ./cmd/l2h-c

clean:
	rm -rf bin/

install:
	go install ./cmd/l2h-s
	go install ./cmd/l2h-c

