package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"l2h/internal/config"
	"l2h/internal/logger"
	"l2h/internal/serverb"
	"l2h/internal/utils"
)

func main() {
	var (
		help          = flag.Bool("help", false, "显示帮助信息")
		showAdminInfo = flag.Bool("show-admin-info", false, "显示管理页面账号密码信息")
		list          = flag.Bool("l", false, "显示当前绑定的路径和端口信息")
		add           = flag.String("a", "", "添加新的路径绑定，格式: path:password")
		delete        = flag.Int("d", -1, "删除某个路径绑定（使用编号）")
		server        = flag.String("s", "", "设置服务器A的地址和API key，格式: server.com:apikey")
		port          = flag.Int("port", 55055, "管理页面端口")
		dataDir       = flag.String("data-dir", "./data", "数据目录")
		daemon        = flag.Bool("daemon", false, "后台运行模式（仅Linux）")
		foreground    = flag.Bool("foreground", false, "强制前台运行")
		pidFile       = flag.String("pid-file", "", "PID文件路径（后台运行时使用）")
	)

	flag.Parse()

	if *help {
		printHelp()
		os.Exit(0)
	}

	// 确保数据目录存在
	if err := os.MkdirAll(*dataDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "创建数据目录失败: %v\n", err)
		os.Exit(1)
	}

	// 数据库文件路径
	dbPath := filepath.Join(*dataDir, "l2h-c.db")

	// 检查是否首次运行
	isFirstRun := !fileExists(dbPath)

	// 如果需要后台运行且不是前台模式
	if *daemon && !*foreground {
		if err := utils.Daemonize(*pidFile); err != nil {
			fmt.Fprintf(os.Stderr, "后台运行失败: %v\n", err)
			os.Exit(1)
		}
		// 守护进程已启动，父进程在这里会退出
		// 子进程继续执行下面的代码
	}

	// 加载配置（简化处理，主要使用命令行参数）
	cfg := config.Default()

	// 初始化日志系统
	logLevel := logger.INFO
	if cfg.ServerB.LogLevel != "" {
		levelMap := map[string]logger.Level{
			"DEBUG": logger.DEBUG,
			"INFO":  logger.INFO,
			"WARN":  logger.WARN,
			"ERROR": logger.ERROR,
			"FATAL": logger.FATAL,
		}
		if level, ok := levelMap[cfg.ServerB.LogLevel]; ok {
			logLevel = level
		}
	}

	var appLogger *logger.Logger
	var err error
	if cfg.ServerB.LogFile != "" {
		appLogger, err = logger.NewFileLogger(logLevel, cfg.ServerB.LogFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "初始化日志系统失败: %v\n", err)
			os.Exit(1)
		}
		defer appLogger.Close()
	} else {
		appLogger = logger.New(logLevel, os.Stdout, "")
	}

	// 如果是首次运行，启动初始化向导
	if isFirstRun {
		fmt.Println()
		if err := runInitWizard(*dataDir, dbPath, *port); err != nil {
			appLogger.Fatal("初始化失败: %v", err)
		}
		// 初始化完成后，如果用户只是想初始化而不是立即启动服务，可以退出
		// 这里我们继续启动服务
	}

	// 确保数据目录存在
	if err := os.MkdirAll(*dataDir, 0755); err != nil {
		appLogger.Fatal("创建数据目录失败: %v", err)
	}

	manager := serverb.NewManager(dbPath)

	if *showAdminInfo {
		info, err := manager.GetAdminInfo()
		if err != nil {
			appLogger.Fatal("获取管理信息失败: %v", err)
		}
		fmt.Printf("管理页面端口: %d\n", *port)
		fmt.Printf("用户名: %s\n", info.Username)
		fmt.Printf("密码: %s\n", info.Password)
		os.Exit(0)
	}

	if *list {
		bindings, err := manager.ListBindings()
		if err != nil {
			appLogger.Fatal("获取绑定列表失败: %v", err)
		}
		if len(bindings) == 0 {
			fmt.Println("当前没有绑定的路径")
		} else {
			fmt.Println("编号\t路径\t\t端口\t\t密码保护")
			for i, binding := range bindings {
				passwordProtected := "否"
				if binding.Password != "" {
					passwordProtected = "是"
				}
				fmt.Printf("%d\t%s\t\t%d\t\t%s\n", i+1, binding.Path, binding.Port, passwordProtected)
			}
		}
		os.Exit(0)
	}

	if *add != "" {
		parts := strings.SplitN(*add, ":", 2)
		if len(parts) < 1 {
			appLogger.Fatal("格式错误，应为 path:password")
		}
		path := parts[0]
		password := ""
		if len(parts) > 1 {
			password = parts[1]
		}

		// 验证路径格式
		if !utils.ValidatePath(path) {
			appLogger.Fatal("路径格式无效，路径不能包含空格或特殊字符，不能以 / 开头或结尾")
		}

		// 检查敏感单词
		if utils.ContainsSensitiveWord(path) {
			appLogger.Fatal("路径包含敏感单词，禁止使用")
		}

		// 提示输入端口
		fmt.Print("请输入端口号: ")
		reader := bufio.NewReader(os.Stdin)
		portStr, err := reader.ReadString('\n')
		if err != nil {
			appLogger.Fatal("读取输入失败: %v", err)
		}
		portStr = strings.TrimSpace(portStr)
		portNum, err := strconv.Atoi(portStr)
		if err != nil {
			appLogger.Fatal("无效的端口号: %v", err)
		}

		// 验证端口号
		if !utils.ValidatePort(portNum) {
			appLogger.Fatal("端口号必须在 1-65535 之间")
		}

		if err := manager.AddBinding(path, portNum, password); err != nil {
			appLogger.Fatal("添加绑定失败: %v", err)
		}
		fmt.Printf("成功添加绑定: %s -> %d\n", path, portNum)
		os.Exit(0)
	}

	if *delete > 0 {
		if err := manager.DeleteBinding(*delete); err != nil {
			appLogger.Fatal("删除绑定失败: %v", err)
		}
		fmt.Printf("成功删除绑定编号: %d\n", *delete)
		os.Exit(0)
	}

	if *server != "" {
		parts := strings.SplitN(*server, ":", 2)
		if len(parts) != 2 {
			appLogger.Fatal("格式错误，应为 server.com:apikey")
		}
		if err := manager.SetServerInfo(parts[0], parts[1]); err != nil {
			appLogger.Fatal("设置服务器信息失败: %v", err)
		}
		fmt.Printf("成功设置服务器信息: %s\n", parts[0])
		os.Exit(0)
	}

	// 如果没有指定任何命令，启动服务
	appLogger.Info("启动服务器B，端口: %d, 数据库: %s", *port, dbPath)
	srv := serverb.NewServer(*port, dbPath)
	if err := srv.Start(); err != nil {
		appLogger.Fatal("启动服务器失败: %v", err)
	}
}

func printHelp() {
	fmt.Println("l2h-c - 服务器B端程序")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  l2h-c [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  --help              显示此帮助信息")
	fmt.Println("  --show-admin-info    显示管理页面账号密码信息")
	fmt.Println("  -l                  显示当前绑定的路径和端口信息")
	fmt.Println("  -a path:password    添加新的路径绑定，password可以为空")
	fmt.Println("  -d <编号>           删除某个路径绑定")
	fmt.Println("  -s server.com:apikey 设置服务器A的地址和API key")
	fmt.Println("  --port              管理页面端口 (默认: 55055)")
	fmt.Println("  --data-dir          数据目录 (默认: ./data)")
	fmt.Println("  --daemon            后台运行模式（仅Linux）")
	fmt.Println("  --foreground        强制前台运行")
	fmt.Println("  --pid-file          PID文件路径（后台运行时使用）")
	fmt.Println()
	fmt.Println("首次运行:")
	fmt.Println("  首次运行时会启动初始化向导，引导您完成基本配置。")
	fmt.Println()
}
