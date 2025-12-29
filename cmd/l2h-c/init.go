package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"

	"l2h/internal/crypto"
	"l2h/internal/serverb"
)

// runInitWizard 运行初始化向导
func runInitWizard(dataDir, dbPath string, defaultPort int) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("========================================")
	fmt.Println("  欢迎使用 l2h-c 客户端")
	fmt.Println("  首次运行初始化向导")
	fmt.Println("========================================")
	fmt.Println()

	// 1. 设置管理员账户
	fmt.Println("1. 设置管理员账户")
	fmt.Print("   请输入管理员用户名 (默认: admin): ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		username = "admin"
	}

	// 设置密码
	fmt.Print("   请输入管理员密码: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("读取密码失败: %w", err)
	}
	fmt.Println()
	password := string(passwordBytes)

	if len(password) < 6 {
		return fmt.Errorf("密码长度至少为 6 个字符")
	}

	fmt.Print("   请再次输入密码: ")
	passwordBytes2, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("读取密码失败: %w", err)
	}
	fmt.Println()
	password2 := string(passwordBytes2)

	if password != password2 {
		return fmt.Errorf("两次输入的密码不一致")
	}

	fmt.Println()

	// 2. 确认数据存储目录
	fmt.Println("2. 数据存储设置")
	fmt.Printf("   数据目录: %s\n", dataDir)
	fmt.Printf("   管理页面端口: %d\n", defaultPort)
	fmt.Println()

	// 3. 可选：设置服务器A信息
	fmt.Println("3. 服务器A配置 (可选)")
	fmt.Print("   是否现在配置服务器A地址和API Key? (y/N): ")
	configServer, _ := reader.ReadString('\n')
	configServer = strings.TrimSpace(strings.ToLower(configServer))

	var serverURL, apiKey string
	if configServer == "y" {
		fmt.Print("   请输入服务器A地址 (例如: example.com): ")
		serverURL, _ = reader.ReadString('\n')
		serverURL = strings.TrimSpace(serverURL)

		fmt.Print("   请输入API Key: ")
		apiKey, _ = reader.ReadString('\n')
		apiKey = strings.TrimSpace(apiKey)
	} else {
		fmt.Println("   跳过服务器A配置，您可以稍后使用 -s 参数配置")
	}

	fmt.Println()

	// 4. 确认配置
	fmt.Println("========================================")
	fmt.Println("  配置预览")
	fmt.Println("========================================")
	fmt.Printf("  管理员用户名: %s\n", username)
	fmt.Printf("  数据目录: %s\n", dataDir)
	fmt.Printf("  管理页面端口: %d\n", defaultPort)
	if serverURL != "" {
		fmt.Printf("  服务器A地址: %s\n", serverURL)
	}
	fmt.Println("========================================")
	fmt.Print("确认以上配置并初始化? (Y/n): ")
	confirm, _ := reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(confirm)) == "n" {
		return fmt.Errorf("用户取消操作")
	}

	fmt.Println()

	// 初始化数据库
	db, err := serverb.NewDatabase(dbPath)
	if err != nil {
		return fmt.Errorf("初始化数据库失败: %w", err)
	}
	defer db.Close()

	// 哈希密码
	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		return fmt.Errorf("加密密码失败: %w", err)
	}

	// 保存管理员账户
	if err := db.SetAdminInfo(username, hashedPassword); err != nil {
		return fmt.Errorf("保存管理员信息失败: %w", err)
	}

	// 如果配置了服务器A信息，也保存
	if serverURL != "" && apiKey != "" {
		if err := db.SetServerInfo(serverURL, apiKey); err != nil {
			return fmt.Errorf("保存服务器信息失败: %w", err)
		}
	}

	fmt.Println("✓ 初始化完成！")
	fmt.Println()
	fmt.Printf("管理页面地址: http://localhost:%d/\n", defaultPort)
	fmt.Println()
	fmt.Println("提示: 您可以使用以下命令管理路径绑定:")
	fmt.Println("  l2h-c -l                    # 查看绑定列表")
	fmt.Println("  l2h-c -a path:password      # 添加新的绑定")
	fmt.Println("  l2h-c -d <编号>             # 删除绑定")
	fmt.Println("  l2h-c --show-admin-info     # 显示管理员信息")
	fmt.Println()

	return nil
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
