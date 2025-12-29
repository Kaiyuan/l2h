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
        # 使用 wget 下载，检查退出码和文件大小
        if ! wget -q "$url" -O "$output" 2>&1; then
            error "下载失败: $url"
        fi
        # wget 在 404 时仍会创建文件（包含错误页面），需要检查
        if [ ! -s "$output" ]; then
            error "下载的文件为空: $output"
        fi
    else
        # 使用 curl 下载，检查 HTTP 状态码
        local http_code=$(curl -sL -o "$output" -w "%{http_code}" "$url")
        if [ "$http_code" != "200" ]; then
            rm -f "$output"
            error "下载失败: $url (HTTP $http_code)。请检查 GitHub Releases 是否存在对应版本。"
        fi
        if [ ! -s "$output" ]; then
            error "下载的文件为空: $output"
        fi
    fi
    
    # 检查文件类型（应该是二进制文件，不应该是文本或HTML）
    if command -v file &> /dev/null; then
        local file_type=$(file -b "$output")
        if [[ "$file_type" == *"text"* ]] || [[ "$file_type" == *"HTML"* ]] || [[ "$file_type" == *"ASCII"* ]]; then
            # 检查文件内容是否包含常见的错误信息
            if head -n 1 "$output" | grep -qE "(Not Found|404|<!DOCTYPE|<html|Repository)"; then
                error "下载的文件是错误页面，不是二进制文件。文件类型: $file_type。请检查 GitHub Releases 是否存在对应版本。"
            fi
            error "下载的文件不是二进制文件: $file_type"
        fi
    fi
    
    # 检查文件大小（二进制文件应该至少几KB）
    local file_size=$(stat -c%s "$output" 2>/dev/null || stat -f%z "$output" 2>/dev/null)
    if [ "$file_size" -lt 1000 ]; then
        error "下载的文件太小，可能不是有效的二进制文件: $output (${file_size} bytes)"
    fi
    
    info "文件下载成功: $output (${file_size} bytes)"
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
    #download_file "${BASE_URL}/l2h-s-${ARCH}.sha256" "l2h-s.sha256"
    
    info "下载 l2h-c..."
    download_file "${BASE_URL}/l2h-c-${ARCH}" "l2h-c"
    #download_file "${BASE_URL}/l2h-c-${ARCH}.sha256" "l2h-c.sha256"
    
    # 验证文件
    #info "验证文件完整性..."
    #verify_file "l2h-s" "l2h-s.sha256"
    #verify_file "l2h-c" "l2h-c.sha256"
    
    # 验证下载的文件是有效的二进制文件
    info "验证下载的文件..."
    if ! command -v file &> /dev/null; then
        warn "无法使用 file 命令验证文件类型，跳过验证"
    else
        local s_type=$(file -b l2h-s)
        local c_type=$(file -b l2h-c)
        if [[ "$s_type" == *"text"* ]] || [[ "$s_type" == *"HTML"* ]] || [[ "$s_type" == *"ASCII"* ]]; then
            error "l2h-s 不是有效的二进制文件: $s_type"
        fi
        if [[ "$c_type" == *"text"* ]] || [[ "$c_type" == *"HTML"* ]] || [[ "$c_type" == *"ASCII"* ]]; then
            error "l2h-c 不是有效的二进制文件: $c_type"
        fi
        info "文件类型验证通过"
    fi
    
    # 安装文件
    info "安装到 $INSTALL_DIR..."
    sudo mkdir -p "$INSTALL_DIR"
    sudo cp l2h-s "$INSTALL_DIR/l2h-s"
    sudo cp l2h-c "$INSTALL_DIR/l2h-c"
    sudo chmod +x "$INSTALL_DIR/l2h-s"
    sudo chmod +x "$INSTALL_DIR/l2h-c"
    
    # 验证安装
    if [ -x "$INSTALL_DIR/l2h-s" ] && [ -x "$INSTALL_DIR/l2h-c" ]; then
        # 尝试运行验证（检查是否是有效的可执行文件）
        if ! "$INSTALL_DIR/l2h-s" --help &>/dev/null && ! "$INSTALL_DIR/l2h-s" --version &>/dev/null; then
            # 如果 --help 和 --version 都失败，检查文件内容
            if head -n 1 "$INSTALL_DIR/l2h-s" | grep -q "Not Found\|404\|<!DOCTYPE\|<html"; then
                error "安装的文件是错误页面，不是有效的二进制文件。请检查 GitHub Releases 是否存在对应版本。"
            fi
        fi
        
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

