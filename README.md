# l2h - 基于 WebRTC 的代理系统

[![Build Status](https://github.com/Kaiyuan/l2h/actions/workflows/build.yml/badge.svg)](https://github.com/Kaiyuan/l2h/actions)
[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

l2h (Local to HTTP) 是一个基于 WebRTC 技术的反向代理系统，通过 WebRTC 连接实现从公网访问内网服务的能力。

## ✨ 特性

- 🔒 **安全**: 使用 Argon2id 加密密码，支持 API Key 认证
- 🌐 **WebRTC**: 基于 WebRTC P2P 连接，无需公网 IP
- 🎯 **灵活**: 支持路径到端口的灵活映射
- 🔑 **权限控制**: 支持无密码访问和密码保护两种模式
- 📊 **管理界面**: 提供基于 PrimeVue V4 的现代化管理后台
- 🚀 **易部署**: 单文件部署，一键安装脚本

## 🏗️ 架构

l2h 由两个主要组件组成：

### 服务器 A (l2h-s)
- 部署在公网服务器上
- 提供 Web 管理界面
- 作为 WebRTC 信令服务器
- 支持路径映射和权限管理
- 生成和管理 API Key

### 服务器 B (l2h-c)
- 运行在内网或本地环境
- 绑定本地端口到访问路径
- 通过 WebRTC 连接到服务器 A
- 转发流量到本地服务

## 📦 快速开始

### 一键安装

#### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/Kaiyuan/l2h/main/install.sh | bash
```

安装脚本会自动检测系统架构并下载对应的二进制文件到 `/usr/local/bin`。

支持的架构：
- x86_64 (Intel/AMD 64位)
- aarch64 (ARM 64位，如树莓派 4/5)
- armv7 (ARM 32位，如树莓派 2/3)

### 手动安装

1. 从 [Releases](https://github.com/Kaiyuan/l2h/releases) 下载对应架构的二进制文件
2. 赋予执行权限：
   ```bash
   chmod +x l2h-s l2h-c
   ```
3. 移动到系统路径：
   ```bash
   sudo mv l2h-s l2h-c /usr/local/bin/
   ```

## 🚀 使用说明

### 服务器 A (l2h-s)

启动服务器 A：

```bash
# 使用默认端口 55080
l2h-s

# 指定端口
l2h-s --port 8080

# 指定数据目录
l2h-s --data-dir /var/lib/l2h

# 使用配置文件
l2h-s --config /etc/l2h/config.json
```

首次启动后访问 `http://your-server:55080` 进行初始化配置。

### 服务器 B (l2h-c)

#### 查看管理信息

```bash
l2h-c --show-admin-info
```

#### 管理路径绑定

```bash
# 列出所有绑定
l2h-c -l

# 添加绑定（会提示输入端口）
l2h-c -a myapp:password123

# 添加无密码绑定
l2h-c -a public-app:

# 删除绑定（使用编号）
l2h-c -d 1
```

#### 设置服务器 A 地址

```bash
l2h-c -s server.example.com:your-api-key
```

#### 启动服务

```bash
# 使用默认端口 55055
l2h-c

# 指定管理端口
l2h-c --port 55055

# 指定数据目录
l2h-c --data-dir /var/lib/l2h
```

启动后可以访问 `http://localhost:55055` 查看管理界面。

### 使用示例

假设您有一个本地运行在 8080 端口的 Web 应用：

1. **在服务器 B 上添加绑定**：
   ```bash
   l2h-c -a myapp:secret123
   # 输入端口: 8080
   ```

2. **在服务器 A 的管理界面配置路径**：
   - 路径: `myapp`
   - 密码: `secret123`
   - 服务器 B 端口: 8080

3. **访问**：
   - 访问 `http://server-a.example.com/myapp`
   - 输入密码 `secret123`
   - 即可访问本地 8080 端口的应用

## 🛠️ 配置文件

### 配置文件格式

创建配置文件 `config.json`：

```json
{
  "server_a": {
    "port": 55080,
    "db_path": "/var/lib/l2h/l2h-s.db",
    "log_file": "/var/log/l2h/l2h-s.log",
    "log_level": "INFO"
  },
  "server_b": {
    "port": 55055,
    "db_path": "/var/lib/l2h/l2h-c.db",
    "log_file": "/var/log/l2h/l2h-c.log",
    "log_level": "INFO"
  },
  "logging": {
    "level": "INFO",
    "stdout": true
  }
}
```

### 日志级别

- `DEBUG`: 调试信息
- `INFO`: 一般信息（默认）
- `WARN`: 警告信息
- `ERROR`: 错误信息
- `FATAL`: 致命错误

## 🔧 从源码编译

### 环境要求

- Go 1.24 或更高版本
- Git

### 克隆仓库

```bash
git clone https://github.com/Kaiyuan/l2h.git
cd l2h
```

### 编译

#### 快速编译（本地平台）

```bash
# 编译所有组件
go build -v ./...

# 单独编译
go build -o bin/l2h-s ./cmd/l2h-s
go build -o bin/l2h-c ./cmd/l2h-c
```

#### 使用 Makefile

```bash
# Linux AMD64 (默认)
make build

# Windows
make build-windows

# macOS
make build-darwin

# 清理构建文件
make clean

# 安装到系统
make install
```

#### 交叉编译

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o bin/l2h-s-linux-amd64 ./cmd/l2h-s

# Linux ARM64 (树莓派 4/5)
GOOS=linux GOARCH=arm64 go build -o bin/l2h-s-linux-arm64 ./cmd/l2h-s

# Linux ARMv7 (树莓派 2/3)
GOOS=linux GOARCH=arm GOARM=7 go build -o bin/l2h-s-linux-armv7 ./cmd/l2h-s

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o bin/l2h-s-windows-amd64.exe ./cmd/l2h-s

# macOS AMD64
GOOS=darwin GOARCH=amd64 go build -o bin/l2h-s-darwin-amd64 ./cmd/l2h-s

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o bin/l2h-s-darwin-arm64 ./cmd/l2h-s
```

### 优化编译

```bash
# 减小二进制文件大小
go build -ldflags="-s -w" -o bin/l2h-s ./cmd/l2h-s

# 添加版本信息
VERSION=$(git describe --tags --always)
go build -ldflags="-s -w -X main.Version=$VERSION" -o bin/l2h-s ./cmd/l2h-s
```

### 运行测试

```bash
# 运行所有测试
go test -v ./...

# 运行代码检查
go vet ./...

# 代码格式化
go fmt ./...
```

## 📁 项目结构

```
l2h/
├── cmd/                    # 命令行入口
│   ├── l2h-s/             # 服务器 A 程序
│   └── l2h-c/             # 服务器 B 程序
├── internal/              # 内部包
│   ├── config/           # 配置管理
│   ├── crypto/           # 加密功能（Argon2id）
│   ├── errors/           # 错误定义
│   ├── logger/           # 日志系统
│   ├── servera/          # 服务器 A 实现
│   │   ├── database.go   # 数据库操作
│   │   ├── server.go     # HTTP 服务器
│   │   └── middleware.go # 中间件
│   ├── serverb/          # 服务器 B 实现
│   │   ├── database.go   # 数据库操作
│   │   ├── server.go     # HTTP 服务器
│   │   └── manager.go    # 管理功能
│   ├── utils/            # 通用工具函数
│   └── webrtc/           # WebRTC 管理
├── .github/              # GitHub Actions
│   └── workflows/
│       └── build.yml     # 自动构建配置
├── install.sh            # 一键安装脚本
├── Makefile             # 构建脚本
├── go.mod               # Go 模块定义
└── README.md            # 项目说明
```

## 🔐 安全性

- **密码加密**: 使用 Argon2id 算法加密存储密码
- **API Key**: 支持 API Key 认证和过期管理
- **路径验证**: 禁止使用敏感词作为路径名
- **输入验证**: 全面的输入参数验证
- **Cookie 安全**: 使用 HttpOnly cookie

## 🤝 贡献

欢迎贡献代码、报告问题或提出建议！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

## 📝 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- [PrimeVue](https://primevue.org/) - 现代化 Vue UI 组件库
- [Go](https://golang.org/) - 高效的编程语言
- [SQLite](https://www.sqlite.org/) - 轻量级数据库

## 📮 联系方式

- 问题反馈: [GitHub Issues](https://github.com/Kaiyuan/l2h/issues)
- 项目主页: [https://github.com/Kaiyuan/l2h](https://github.com/Kaiyuan/l2h)

---

**注意**: 本项目仍在开发中，功能可能会有变动。建议在生产环境使用前进行充分测试。
