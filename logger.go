package gone

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// GLogger 日志
type GLogger struct {
	writer io.Writer
	lock   sync.Mutex
}

// newLogger GLogger构造函数
func newLogger(w io.Writer) *GLogger {
	return &GLogger{
		writer: w,
		lock:   sync.Mutex{},
	}
}

// SetStatus 设置 GLogger 状态，控制输出
func (logger *GLogger) SetStatus(value bool) {
	logger.lock.Lock()
	defer logger.lock.Unlock()
	if value {
		logger.writer = os.Stdout
	} else {
		logger.writer = io.Discard
	}
}

// Println 行输出
func (logger *GLogger) Println(v ...any) {
	logger.writer.Write([]byte(fmt.Sprintln(v...)))
}

// Printf 格式化输出
func (logger *GLogger) Printf(format string, v ...any) {
	logger.writer.Write([]byte(fmt.Sprintf(format, v...)))
}

var (
	loggerInfo    = newLogger(os.Stdout)
	loggerWarn    = newLogger(os.Stdout)
	loggerError   = newLogger(os.Stderr)
	loggerWatcher = newLogger(io.Discard)
)

var (
	LogInfo   = info
	LogInfof  = infof
	LogWarn   = warn
	LogWarnf  = warnf
	LogError  = err
	LogErrorf = errf
	logWatchf = watchPrtf
)

func info(v ...any) {
	v = append([]any{"\033[1;36m", "[INFO]", time.Now().Format("2006-01-02 15:04:05")}, v...)
	v = append(v, "\033[0m")
	loggerInfo.Println(v...)
}

func infof(format string, v ...any) {
	info(fmt.Sprintf(format, v...))
}

func warn(v ...any) {
	v = append([]any{"\033[1;33m", "[WARN]", time.Now().Format("2006-01-02 15:04:05")}, v...)
	v = append(v, "\033[0m")
	loggerWarn.Println(v...)
}

func warnf(format string, v ...any) {
	warn(fmt.Sprintf(format, v...))
}

func err(v ...any) {
	v = append([]any{"\033[1;31m", "[ERROR]", time.Now().Format("2006-01-02 15:04:05")}, v...)
	v = append(v, "\033[0m")
	loggerError.Println(v...)
}

func errf(format string, v ...any) {
	err(fmt.Sprintf(format, v...))
}

func watchPrt(v ...any) {
	v = append([]any{"\033[1;35m", "[WATCHER]"}, v...)
	v = append(v, "\033[0m")
	loggerWatcher.Println(v...)
}

func watchPrtf(format string, v ...any) {
	watchPrt(fmt.Sprintf(format, v...))
}

var spinners = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

func GetSpinner() func() string {
	var index int = 0
	return func() string {
		index = (index + 1) % len(spinners)
		return spinners[index]
	}
}
