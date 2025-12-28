# GitHub Actions 工作流说明

## 自动构建

本项目使用 GitHub Actions 自动构建 Linux 二进制文件，支持以下架构：

- **x86_64** (amd64)
- **aarch64** (arm64)
- **armv7** (arm)

## 触发条件

工作流会在以下情况自动运行：

1. **推送标签**：当推送以 `v` 开头的标签时（如 `v1.0.0`），会自动构建并创建 Release
2. **Pull Request**：在 PR 中会构建所有架构的二进制文件用于测试
3. **手动触发**：可以在 GitHub Actions 页面手动触发工作流

## 构建产物

每次构建会生成：

- `l2h-s-{arch}` - 服务器A程序
- `l2h-c-{arch}` - 服务器B程序
- `l2h-s-{arch}.sha256` - 服务器A程序校验文件
- `l2h-c-{arch}.sha256` - 服务器B程序校验文件

## 创建 Release

当推送标签时，工作流会自动：

1. 构建所有架构的二进制文件
2. 生成 SHA256 校验文件
3. 创建 GitHub Release
4. 上传所有构建产物到 Release

## 使用方法

### 创建新版本

```bash
# 1. 更新版本号（如果需要）
# 2. 提交更改
git add .
git commit -m "Release v1.0.0"
# 3. 创建标签
git tag v1.0.0
# 4. 推送代码和标签
git push origin main
git push origin v1.0.0
```

### 手动触发构建

1. 访问 GitHub 仓库的 Actions 页面
2. 选择 "Build Linux Binaries" 工作流
3. 点击 "Run workflow"
4. 选择分支并运行

## 注意事项

- 确保 Go 版本为 1.21 或更高
- Release 创建需要仓库的写入权限
- 构建产物会保留 90 天（GitHub Actions 默认设置）

