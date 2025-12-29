# L2H 安装说明

## 一键安装（推荐）

使用一键安装脚本自动下载并安装最新版本：

```bash
curl -fsSL https://raw.githubusercontent.com/Kaiyuan/l2h/main/install.sh | bash
```

或者使用 wget：

```bash
wget -qO- https://raw.githubusercontent.com/Kaiyuan/l2h/main/install.sh | bash
```

## 手动安装

### 1. 下载对应架构的二进制文件

访问 [Releases 页面](https://github.com/Kaiyuan/l2h/releases) 下载对应架构的二进制文件：

- **x86_64 (amd64)**: `l2h-s-x86_64`, `l2h-c-x86_64`
- **aarch64 (arm64)**: `l2h-s-aarch64`, `l2h-c-aarch64`
- **armv7**: `l2h-s-armv7`, `l2h-c-armv7`

### 2. 验证文件完整性

下载对应的 SHA256 校验文件并验证：

```bash
# 下载校验文件
wget https://github.com/Kaiyuan/l2h/releases/download/v1.0.0/l2h-s-x86_64.sha256
wget https://github.com/Kaiyuan/l2h/releases/download/v1.0.0/l2h-c-x86_64.sha256

# 验证
sha256sum -c l2h-s-x86_64.sha256
sha256sum -c l2h-c-x86_64.sha256
```

### 3. 安装到系统

```bash
# 复制到系统路径
sudo cp l2h-s-x86_64 /usr/local/bin/l2h-s
sudo cp l2h-c-x86_64 /usr/local/bin/l2h-c

# 添加执行权限
sudo chmod +x /usr/local/bin/l2h-s
sudo chmod +x /usr/local/bin/l2h-c
```

### 4. 验证安装

```bash
l2h-s --help
l2h-c --help
```

## 从源码编译

### 前置要求

- Go 1.21 或更高版本

### 编译步骤

```bash
# 克隆仓库
git clone https://github.com/Kaiyuan/l2h.git
cd l2h

# 编译 Linux 版本（默认）
make build

# 或指定架构
GOOS=linux GOARCH=amd64 go build -o bin/l2h-s ./cmd/l2h-s
GOOS=linux GOARCH=amd64 go build -o bin/l2h-c ./cmd/l2h-c
```

## 系统要求

- Linux 系统（支持 x86_64、aarch64、armv7）
- 足够的磁盘空间（约 10MB）

## 卸载

```bash
sudo rm /usr/local/bin/l2h-s
sudo rm /usr/local/bin/l2h-c
```

## 故障排除

### 安装脚本失败

如果一键安装脚本失败，请检查：

1. 网络连接是否正常
2. 是否有 sudo 权限
3. 系统架构是否支持

### 权限问题

如果遇到权限问题，可以安装到用户目录：

```bash
export INSTALL_DIR=$HOME/.local/bin
curl -fsSL https://raw.githubusercontent.com/Kaiyuan/l2h/main/install.sh | bash
```

然后添加到 PATH：

```bash
echo 'export PATH=$HOME/.local/bin:$PATH' >> ~/.bashrc
source ~/.bashrc
```

### "Not: command not found" 错误

如果运行 `l2h-s` 或 `l2h-c` 时出现 `Not: command not found` 错误，这通常表示：

1. **下载的文件是错误页面**：GitHub Releases 中可能没有对应版本的二进制文件
2. **网络问题**：下载过程中出现了错误，但脚本没有正确检测到

**解决方法**：

1. **检查是否有可用的 Release**：
   ```bash
   curl -s https://api.github.com/repos/Kaiyuan/l2h/releases/latest
   ```
   如果返回 404，说明还没有创建任何 Release。

2. **从源码编译**（推荐）：
   ```bash
   git clone https://github.com/Kaiyuan/l2h.git
   cd l2h
   make build
   sudo cp bin/l2h-s /usr/local/bin/l2h-s
   sudo cp bin/l2h-c /usr/local/bin/l2h-c
   sudo chmod +x /usr/local/bin/l2h-s
   sudo chmod +x /usr/local/bin/l2h-c
   ```

3. **手动下载并验证**：
   ```bash
   # 检查下载的文件类型
   file /usr/local/bin/l2h-s
   # 如果显示 "ASCII text" 或 "HTML"，说明下载的是错误页面
   # 需要删除并重新安装
   sudo rm /usr/local/bin/l2h-s /usr/local/bin/l2h-c
   ```

## 支持

如有问题，请访问 [GitHub Issues](https://github.com/Kaiyuan/l2h/issues)

