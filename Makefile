.PHONY: build build-windows build-linux build-darwin clean install frontend

frontend:
	@if ! command -v npm >/dev/null 2>&1; then \
		echo "Error: npm is not installed. Please install Node.js/npm to build the frontend."; \
		exit 1; \
	fi
	cd web/admin && npm install && npm run build
	rm -rf internal/servera/static
	mkdir -p internal/servera/static
	cp -r web/admin/dist/* internal/servera/static/

# 默认构建 Linux 版本（主要使用）
build: frontend
	GOOS=linux GOARCH=amd64 go build -o bin/l2h-s ./cmd/l2h-s
	GOOS=linux GOARCH=amd64 go build -o bin/l2h-c ./cmd/l2h-c

build-windows: frontend
	go build -o bin/l2h-s.exe ./cmd/l2h-s
	go build -o bin/l2h-c.exe ./cmd/l2h-c

build-linux: frontend
	GOOS=linux GOARCH=amd64 go build -o bin/l2h-s ./cmd/l2h-s
	GOOS=linux GOARCH=amd64 go build -o bin/l2h-c ./cmd/l2h-c

build-darwin: frontend
	GOOS=darwin GOARCH=amd64 go build -o bin/l2h-s ./cmd/l2h-s
	GOOS=darwin GOARCH=amd64 go build -o bin/l2h-c ./cmd/l2h-c

clean:
	rm -rf bin/
	rm -rf internal/servera/static

install: frontend
	go install ./cmd/l2h-s
	go install ./cmd/l2h-c

