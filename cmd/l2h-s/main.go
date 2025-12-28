package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

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
		log.Fatalf("创建数据目录失败: %v", err)
	}

	// 数据库文件路径
	dbPath := filepath.Join(*dataDir, "l2h-s.db")

	server := servera.NewServer(*port, dbPath, *configFile)
	if err := server.Start(); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
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
