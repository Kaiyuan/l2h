# 构建说明

## 依赖要求

- Go 1.21 或更高版本
- Node.js 18 或更高版本 (用于构建前端)

## 构建步骤

### Linux（主要使用）

```bash
make build
```

或者直接使用 go build:

```bash
go build -o bin/l2h-s ./cmd/l2h-s
go build -o bin/l2h-c ./cmd/l2h-c
```

### Windows

```bash
make build-windows
```

### macOS

```bash
make build-darwin
```

## 安装到系统

```bash
make install
```

安装后，`l2h-s` 和 `l2h-c` 将可以在任何地方使用。

## 使用说明

### 服务器A (l2h-s)

```bash
# 查看帮助
./bin/l2h-s --help

# 启动服务器（默认端口 55080）
./bin/l2h-s

# 或指定端口
./bin/l2h-s --port 55080
```

### 服务器B (l2h-c)

```bash
# 查看帮助
./bin/l2h-c --help

# 显示管理信息
./bin/l2h-c --show-admin-info

# 列出所有绑定
./bin/l2h-c -l

# 添加绑定
./bin/l2h-c -a pathB:password

# 删除绑定（编号从1开始）
./bin/l2h-c -d 1

# 设置服务器A信息
./bin/l2h-c -s server.com:apikey

# 启动服务器（默认管理端口 55055）
./bin/l2h-c
```

