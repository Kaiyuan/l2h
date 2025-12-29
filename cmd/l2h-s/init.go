package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/term"

	"l2h/internal/config"
	"l2h/internal/crypto"
	"l2h/internal/servera"
	"l2h/internal/utils"
)

// runInitWizard 运行初始化向导
func runInitWizard(dataDir, configPath string, defaultPort int) (*config.Config, error) {
	reader := bufio.NewReader(os.Stdin)

	// 1. 设置管理页面 URL 目录
	fmt.Println("1. 设置管理页面访问路径")
	fmt.Print("   请输入管理页面的 URL 路径 (默认: admin): ")
	adminPath, _ := reader.ReadString('\n')
	adminPath = strings.TrimSpace(adminPath)
	if adminPath == "" {
		adminPath = "admin"
	}

	// 验证路径格式
	if !utils.ValidatePath(adminPath) {
		return nil, fmt.Errorf("无效的路径格式: %s", adminPath)
	}

	// 检查敏感词
	if utils.ContainsSensitiveWord(adminPath) {
		fmt.Println("   警告: 该路径包含敏感词，建议使用其他路径名")
		fmt.Print("   确定要继续使用这个路径吗? (y/N): ")
		confirm, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
			return nil, fmt.Errorf("用户取消操作")
		}
	}

	fmt.Println()

	// 2. 确认数据存储目录
	fmt.Println("2. 设置数据存储目录")
	fmt.Printf("   数据目录: %s\n", dataDir)
	fmt.Println("   (数据库和配置文件将保存在此目录)")
	fmt.Println()

	// 3. 设置管理员账户
	fmt.Println("3. 设置管理员账户")
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
		return nil, fmt.Errorf("读取密码失败: %w", err)
	}
	fmt.Println()
	password := string(passwordBytes)

	if len(password) < 6 {
		return nil, fmt.Errorf("密码长度至少为 6 个字符")
	}

	fmt.Print("   请再次输入密码: ")
	passwordBytes2, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, fmt.Errorf("读取密码失败: %w", err)
	}
	fmt.Println()
	password2 := string(passwordBytes2)

	if password != password2 {
		return nil, fmt.Errorf("两次输入的密码不一致")
	}

	// 4. 可选：设置邮箱
	fmt.Print("   请输入管理员邮箱 (可选，直接回车跳过): ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	if email != "" && !utils.ValidateEmail(email) {
		fmt.Println("   警告: 邮箱格式可能不正确，但将继续保存")
	}

	fmt.Println()

	// 5. 确认配置
	fmt.Println("========================================")
	fmt.Println("  配置预览")
	fmt.Println("========================================")
	fmt.Printf("  管理页面路径: /%s\n", adminPath)
	fmt.Printf("  数据目录: %s\n", dataDir)
	fmt.Printf("  管理员用户名: %s\n", username)
	fmt.Printf("  管理员邮箱: %s\n", email)
	fmt.Printf("  服务器端口: %d\n", defaultPort)
	fmt.Println("========================================")
	fmt.Print("确认以上配置并初始化? (Y/n): ")
	confirm, _ := reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(confirm)) == "n" {
		return nil, fmt.Errorf("用户取消操作")
	}

	fmt.Println()

	// 创建配置
	cfg := &config.Config{
		ServerA: config.ServerAConfig{
			Port:     defaultPort,
			DBPath:   filepath.Join(dataDir, "l2h-s.db"),
			LogFile:  "l2h-s.log",
			LogLevel: "INFO",
		},
		Logging: config.LoggingConfig{
			Level:  "INFO",
			Stdout: true,
		},
	}

	// 保存配置文件
	if err := config.Save(configPath, cfg); err != nil {
		return nil, fmt.Errorf("保存配置文件失败: %w", err)
	}

	// 初始化数据库并保存设置
	dbPath := cfg.ServerA.DBPath
	db, err := servera.NewDatabase(dbPath)
	if err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %w", err)
	}

	// 哈希密码
	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("加密密码失败: %w", err)
	}

	// 保存管理员设置
	settings := &servera.Settings{
		AdminPath: adminPath,
		Username:  username,
		Password:  hashedPassword,
		Email:     email,
	}

	if err := db.SetSettings(settings); err != nil {
		return nil, fmt.Errorf("保存管理员设置失败: %w", err)
	}

	db.Close()

	fmt.Println("✓ 初始化完成！")
	fmt.Println()
	fmt.Printf("管理页面地址: http://localhost:%d/%s\n", defaultPort, adminPath)
	fmt.Println()

	return cfg, nil
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func printHelp() {
	fmt.Println("l2h-s - 服务器A端程序")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  l2h-s [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  --help          显示此帮助信息")
	fmt.Println("  --port          服务器端口 (默认: 55080)")
	fmt.Println("  --data-dir      数据目录 (默认: ./data)")
	fmt.Println("  --config        配置文件路径")
	fmt.Println("  --daemon        后台运行模式（仅Linux）")
	fmt.Println("  --foreground    强制前台运行")
	fmt.Println("  --pid-file      PID文件路径（后台运行时使用）")
	fmt.Println()
	fmt.Println("首次运行:")
	fmt.Println("  首次运行时会启动初始化向导，引导您完成基本配置。")
	fmt.Println("  配置将保存在数据目录中的 config.json 文件里。")
	fmt.Println()
}
