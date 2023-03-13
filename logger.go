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
	name   string
	writer io.Writer
	color  int
	prefix string
	lock   sync.Mutex
}

// NewGLogger 新建logger
func NewGLogger(name string, output io.Writer, prefix string, color int) *GLogger {
	logger := &GLogger{
		name:   name,
		writer: output,
		color:  color,
		prefix: prefix,
		lock:   sync.Mutex{},
	}
	gloggers[name] = logger
	return logger
}

const (
	LogRed    int = 31
	LogGreen  int = 32
	LogYellow int = 33
	LogBlue   int = 36
	LogPurple int = 35
)

var (
	infoLogger  = NewGLogger("info", os.Stdout, "[INFO]", LogBlue)
	warnLogger  = NewGLogger("warning", os.Stdout, "[WARNING]", LogYellow)
	errLogger   = NewGLogger("error", os.Stdout, "[ERROR]", LogRed)
	watchLogger = NewGLogger("watch", io.Discard, "[WATCH]", LogPurple)
	gloggers    = make(map[string]*GLogger)
)

// Printf 格式化输出
func (logger *GLogger) Printf(format string, v ...any) {
	value := fmt.Sprintf(format, v...)
	value = fmt.Sprintf("\033[1;%dm%s %s %s\033[0m", logger.color, logger.prefix, time.Now().Format("2006-01-02 15:04:05"), value)
	logger.writer.Write([]byte(value))
}

// Println 按行输出
func (logger *GLogger) Println(v ...any) {
	value := fmt.Sprint(v...)
	value = fmt.Sprintf("\033[1;%dm%s %s %s\033[0m\n", logger.color, logger.prefix, time.Now().Format("2006-01-02 15:04:05"), value)
	logger.writer.Write([]byte(value))
}

// SetStatus 设置输出状态
func (logger *GLogger) SetStatus(value bool) {
	logger.lock.Lock()
	defer logger.lock.Unlock()
	if value {
		logger.writer = os.Stdout
	} else {
		logger.writer = io.Discard
	}
}

var (
	LogInfo   = infoLogger.Println
	LogInfof  = infoLogger.Printf
	LogWarn   = warnLogger.Println
	LogWarnf  = warnLogger.Printf
	LogErr    = errLogger.Println
	LogErrf   = errLogger.Printf
	logWatch  = watchLogger.Println
	logWatchf = watchLogger.Printf
)

// 加载动画
var spinners = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

// Spinner 获取加载图标
func Spinner() func() string {
	var index int = 0
	return func() string {
		index = (index + 1) % len(spinners)
		return spinners[index]
	}
}
