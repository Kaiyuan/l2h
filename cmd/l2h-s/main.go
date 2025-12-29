package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"l2h/internal/config"
	"l2h/internal/logger"
	"l2h/internal/servera"
)

func main() {
	var (
		help       = flag.Bool("help", false, "显示帮助信息")
		port       = flag.Int("port", 55080, "服务器端口")
		dataDir    = flag.String("data-dir", "./data", "数据目录")
		configFile = flag.String("config", "", "配置文件路径")
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

	// 配置文件路径
	configPath := *configFile
	if configPath == "" {
		configPath = filepath.Join(*dataDir, "config.json")
	}

	// 数据库文件路径
	dbPath := filepath.Join(*dataDir, "l2h-s.db")

	// 检查是否首次运行
	isFirstRun := !fileExists(dbPath) && !fileExists(configPath)

	var cfg *config.Config
	var err error

	if isFirstRun {
		// 首次运行，启动初始化向导
		fmt.Println("========================================")
		fmt.Println("  欢迎使用 l2h-s 服务器")
		fmt.Println("  首次运行初始化向导")
		fmt.Println("========================================")
		fmt.Println()

		cfg, err = runInitWizard(*dataDir, configPath, *port)
		if err != nil {
			fmt.Fprintf(os.Stderr, "初始化失败: %v\n", err)
			os.Exit(1)
		}
	} else {
		// 加载现有配置
		cfg, err = config.Load(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载配置文件失败: %v\n", err)
			os.Exit(1)
		}
	}

	// 初始化日志系统
	logLevel := logger.INFO
	if cfg.ServerA.LogLevel != "" {
		levelMap := map[string]logger.Level{
			"DEBUG": logger.DEBUG,
			"INFO":  logger.INFO,
			"WARN":  logger.WARN,
			"ERROR": logger.ERROR,
			"FATAL": logger.FATAL,
		}
		if level, ok := levelMap[cfg.ServerA.LogLevel]; ok {
			logLevel = level
		}
	}

	var logFile string
	if cfg.ServerA.LogFile != "" {
		logFile = filepath.Join(*dataDir, cfg.ServerA.LogFile)
	}

	appLogger, err := logger.NewFileLogger(logLevel, logFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志系统失败: %v\n", err)
		os.Exit(1)
	}
	defer appLogger.Close()

	// 使用配置或命令行参数
	serverPort := *port
	if cfg.ServerA.Port > 0 {
		serverPort = cfg.ServerA.Port
	}

	// 更新配置中的数据库路径
	if cfg.ServerA.DBPath == "" || !filepath.IsAbs(cfg.ServerA.DBPath) {
		cfg.ServerA.DBPath = dbPath
	}

	appLogger.Info("启动服务器A，端口: %d, 数据库: %s", serverPort, cfg.ServerA.DBPath)

	if isFirstRun {
		appLogger.Info("首次运行，配置已保存到: %s", configPath)
	}

	server := servera.NewServer(serverPort, cfg.ServerA.DBPath, configPath)
	if err := server.Start(); err != nil {
		appLogger.Fatal("启动服务器失败: %v", err)
	}
}
