//go:build windows
// +build windows

package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Daemonize Windows 不支持真正的守护进程
// 返回友好的错误信息，建议用户使用其他方法
func Daemonize(pidFile string) error {
	return fmt.Errorf("Windows 不支持守护进程模式\n" +
		"请使用以下方法之一在后台运行:\n" +
		"1. 使用 PowerShell: Start-Process -NoNewWindow -FilePath \"程序路径\" -ArgumentList \"参数\"\n" +
		"2. 使用 cmd: start /b 程序路径 参数\n" +
		"3. 创建 Windows 服务\n" +
		"4. 使用任务计划程序")
}

// writePidFile 写入当前进程的 PID 到文件（Windows 也支持）
func writePidFile(pidFile string) error {
	if pidFile == "" {
		return nil
	}

	// 确保 PID 文件目录存在
	dir := filepath.Dir(pidFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建PID文件目录失败: %w", err)
	}

	// 写入 PID
	pidStr := strconv.Itoa(os.Getpid())
	if err := os.WriteFile(pidFile, []byte(pidStr+"\r\n"), 0644); err != nil {
		return fmt.Errorf("写入PID文件失败: %w", err)
	}

	return nil
}

// RemovePidFile 删除 PID 文件
func RemovePidFile(pidFile string) error {
	if pidFile == "" {
		return nil
	}
	return os.Remove(pidFile)
}
