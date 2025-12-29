//go:build linux || darwin
// +build linux darwin

package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

// Daemonize 将当前进程转换为守护进程（仅支持 Unix-like 系统）
// pidFile: PID 文件路径，如果为空则不创建 PID 文件
func Daemonize(pidFile string) error {
	// 检查是否已经是守护进程
	if os.Getppid() == 1 {
		// 已经是守护进程，写入 PID 文件并返回
		if pidFile != "" {
			return writePidFile(pidFile)
		}
		return nil
	}

	// 获取当前可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	// 准备启动参数（排除 --daemon 标志以避免无限循环）
	args := make([]string, 0, len(os.Args)-1)
	skipNext := false
	for i, arg := range os.Args {
		if i == 0 {
			continue // 跳过程序名
		}
		if skipNext {
			skipNext = false
			continue
		}
		// 过滤掉 --daemon 相关的参数
		if arg == "--daemon" || arg == "-daemon" {
			continue
		}
		if arg == "--foreground" || arg == "-foreground" {
			continue
		}
		args = append(args, arg)
	}

	// 添加 --foreground 标志，确保子进程不会再次尝试守护化
	args = append(args, "--foreground")

	// 启动子进程
	cmd := exec.Command(execPath, args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // 创建新的会话
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动守护进程失败: %w", err)
	}

	// 写入 PID 文件（在子进程中）
	if pidFile != "" {
		// 父进程退出，子进程写入自己的 PID
		childPid := cmd.Process.Pid
		if err := writePidFileWithPid(pidFile, childPid); err != nil {
			// 即使失败也不阻止进程继续
			fmt.Fprintf(os.Stderr, "警告: 写入PID文件失败: %v\n", err)
		}
	}

	// 父进程退出
	os.Exit(0)
	return nil
}

// writePidFile 写入当前进程的 PID 到文件
func writePidFile(pidFile string) error {
	return writePidFileWithPid(pidFile, os.Getpid())
}

// writePidFileWithPid 写入指定 PID 到文件
func writePidFileWithPid(pidFile string, pid int) error {
	// 确保 PID 文件目录存在
	dir := filepath.Dir(pidFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建PID文件目录失败: %w", err)
	}

	// 写入 PID
	pidStr := strconv.Itoa(pid)
	if err := os.WriteFile(pidFile, []byte(pidStr+"\n"), 0644); err != nil {
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
