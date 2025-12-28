#!/bin/bash

# L2H 一键安装脚本
# GitHub: https://github.com/Kaiyuan/l2h

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检测系统架构
detect_arch() {
    local arch=$(uname -m)
    case "$arch" in
        x86_64)
            echo "x86_64"
            ;;
        aarch64|arm64)
            echo "aarch64"
            ;;
        armv7l|armv6l|arm)
            echo "armv7"
            ;;
        *)
            echo "unsupported"
            ;;
    esac
}

# 检测操作系统
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        echo "$ID"
    elif [ -f /etc/redhat-release ]; then
        echo "rhel"
    else
        echo "unknown"
    fi
}

# 打印信息
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# 检查命令是否存在
check_command() {
    if ! command -v "$1" &> /dev/null; then
        return 1
    fi
    return 0
}

# 安装依赖
install_dependencies() {
    local os=$(detect_os)
    info "检测到操作系统: $os"
    
    if check_command wget; then
        DOWNLOAD_CMD="wget -q"
    elif check_command curl; then
        DOWNLOAD_CMD="curl -sL"
    else
        error "需要 wget 或 curl 来下载文件"
    fi
    
    info "使用下载工具: $DOWNLOAD_CMD"
}

# 获取最新版本
get_latest_version() {
    if check_command jq; then
        local version=$(curl -s https://api.github.com/repos/Kaiyuan/l2h/releases/latest | jq -r '.tag_name')
        echo "${version#v}"
    else
        local version=$(curl -s https://api.github.com/repos/Kaiyuan/l2h/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | head -1)
        echo "${version#v}"
    fi
}

# 下载文件
download_file() {
    local url=$1
    local output=$2
    
    if [[ "$DOWNLOAD_CMD" == *"wget"* ]]; then
        wget -q "$url" -O "$output" || error "下载失败: $url"
    else
        curl -sL "$url" -o "$output" || error "下载失败: $url"
    fi
}

# 验证文件
verify_file() {
    local file=$1
    local sha256_file=$2
    
    if [ ! -f "$sha256_file" ]; then
        warn "SHA256 文件不存在，跳过验证"
        return 0
    fi
    
    local expected=$(cat "$sha256_file" | awk '{print $1}')
    local actual=$(sha256sum "$file" | awk '{print $1}')
    
    if [ "$expected" != "$actual" ]; then
        error "文件校验失败"
    fi
    
    info "文件校验通过"
}

# 主安装函数
main() {
    echo "=========================================="
    echo "  L2H 一键安装脚本"
    echo "  GitHub: https://github.com/Kaiyuan/l2h"
    echo "=========================================="
    echo ""
    
    # 检测架构
    ARCH=$(detect_arch)
    if [ "$ARCH" == "unsupported" ]; then
        error "不支持的架构: $(uname -m)"
    fi
    info "检测到架构: $ARCH"
    
    # 安装依赖
    install_dependencies
    
    # 获取版本
    info "获取最新版本..."
    VERSION=$(get_latest_version)
    if [ -z "$VERSION" ]; then
        error "无法获取版本信息"
    fi
    info "最新版本: v$VERSION"
    
    # 设置安装目录
    INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
    info "安装目录: $INSTALL_DIR"
    
    # 创建临时目录
    TEMP_DIR=$(mktemp -d)
    trap "rm -rf $TEMP_DIR" EXIT
    
    cd "$TEMP_DIR"
    
    # 下载文件
    BASE_URL="https://github.com/Kaiyuan/l2h/releases/download/v${VERSION}"
    
    info "下载 l2h-s..."
    download_file "${BASE_URL}/l2h-s-${ARCH}" "l2h-s"
    download_file "${BASE_URL}/l2h-s-${ARCH}.sha256" "l2h-s.sha256"
    
    info "下载 l2h-c..."
    download_file "${BASE_URL}/l2h-c-${ARCH}" "l2h-c"
    download_file "${BASE_URL}/l2h-c-${ARCH}.sha256" "l2h-c.sha256"
    
    # 验证文件
    info "验证文件完整性..."
    verify_file "l2h-s" "l2h-s.sha256"
    verify_file "l2h-c" "l2h-c.sha256"
    
    # 安装文件
    info "安装到 $INSTALL_DIR..."
    sudo mkdir -p "$INSTALL_DIR"
    sudo cp l2h-s "$INSTALL_DIR/l2h-s"
    sudo cp l2h-c "$INSTALL_DIR/l2h-c"
    sudo chmod +x "$INSTALL_DIR/l2h-s"
    sudo chmod +x "$INSTALL_DIR/l2h-c"
    
    # 验证安装
    if [ -x "$INSTALL_DIR/l2h-s" ] && [ -x "$INSTALL_DIR/l2h-c" ]; then
        echo ""
        echo "=========================================="
        info "安装成功！"
        echo "=========================================="
        echo ""
        echo "已安装:"
        echo "  - l2h-s: $INSTALL_DIR/l2h-s"
        echo "  - l2h-c: $INSTALL_DIR/l2h-c"
        echo ""
        echo "使用说明:"
        echo "  l2h-s --help    # 查看服务器A帮助"
        echo "  l2h-c --help    # 查看服务器B帮助"
        echo ""
    else
        error "安装验证失败"
    fi
}

# 运行主函数
main

