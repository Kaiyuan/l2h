package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelNames = map[Level]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

type Logger struct {
	level  Level
	logger *log.Logger
	file   *os.File
}

var defaultLogger *Logger

func init() {
	defaultLogger = New(INFO, os.Stdout, "")
}

// New 创建新的日志记录器
func New(level Level, output io.Writer, logFile string) *Logger {
	var writers []io.Writer
	writers = append(writers, output)

	var file *os.File
	if logFile != "" {
		// 确保日志目录存在
		dir := filepath.Dir(logFile)
		if err := os.MkdirAll(dir, 0755); err == nil {
			f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err == nil {
				file = f
				writers = append(writers, file)
			}
		}
	}

	multiWriter := io.MultiWriter(writers...)
	logger := log.New(multiWriter, "", log.LstdFlags)

	return &Logger{
		level:  level,
		logger: logger,
		file:   file,
	}
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level Level) {
	l.level = level
}

// Close 关闭日志文件
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	prefix := fmt.Sprintf("[%s] ", levelNames[level])
	message := fmt.Sprintf(format, args...)
	l.logger.Printf("%s%s", prefix, message)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
	os.Exit(1)
}

// 全局日志函数
func SetLevel(level Level) {
	defaultLogger.SetLevel(level)
}

func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	defaultLogger.Fatal(format, args...)
}

// NewFileLogger 创建文件日志记录器
func NewFileLogger(level Level, logFile string) (*Logger, error) {
	dir := filepath.Dir(logFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 添加时间戳到日志文件名
	ext := filepath.Ext(logFile)
	base := logFile[:len(logFile)-len(ext)]
	timestamp := time.Now().Format("20060102-150405")
	logFile = fmt.Sprintf("%s-%s%s", base, timestamp, ext)

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %w", err)
	}

	multiWriter := io.MultiWriter(os.Stdout, file)
	logger := log.New(multiWriter, "", log.LstdFlags)

	return &Logger{
		level:  level,
		logger: logger,
		file:   file,
	}, nil
}

