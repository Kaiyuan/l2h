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

	// 加载配置
	var cfg *config.Config
	var err error
	if *configFile != "" {
		cfg, err = config.Load(*configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载配置文件失败: %v\n", err)
			os.Exit(1)
		}
	} else {
		cfg = config.Default()
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
		logFile = cfg.ServerA.LogFile
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

	dataDirPath := *dataDir
	if cfg.ServerA.DBPath != "" {
		dataDirPath = filepath.Dir(cfg.ServerA.DBPath)
	}

	// 确保数据目录存在
	if err := os.MkdirAll(dataDirPath, 0755); err != nil {
		appLogger.Fatal("创建数据目录失败: %v", err)
	}

	// 数据库文件路径
	dbPath := cfg.ServerA.DBPath
	if dbPath == "" {
		dbPath = filepath.Join(dataDirPath, "l2h-s.db")
	}

	appLogger.Info("启动服务器A，端口: %d, 数据库: %s", serverPort, dbPath)

	server := servera.NewServer(serverPort, dbPath, *configFile)
	if err := server.Start(); err != nil {
		appLogger.Fatal("启动服务器失败: %v", err)
	}
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
	fmt.Println()
}
