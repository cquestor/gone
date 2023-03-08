package gone

import (
	"fmt"
	"log"
	"os"
)

// TODO: 自定义Logger
var (
	loggerInfo  = log.New(os.Stderr, "[INFO]", log.LstdFlags)
	loggerWarn  = log.New(os.Stderr, "[WARN]", log.LstdFlags)
	loggerError = log.New(os.Stderr, "[ERROR]", log.LstdFlags)
)

var (
	LogInfo   = info
	LogInfof  = infof
	LogWarn   = warn
	LogWarnf  = warnf
	LogError  = err
	LogErrorf = errf
)

func info(v ...any) {
	v = append([]any{"\033[1;36m"}, v...)
	v = append(v, "\033[0m")
	loggerInfo.Println(v...)
}

func infof(format string, v ...any) {
	info(fmt.Sprintf(format, v...))
}

func warn(v ...any) {
	v = append([]any{"\033[1;33m"}, v...)
	v = append(v, "\033[0m")
	loggerWarn.Println(v...)
}

func warnf(format string, v ...any) {
	warn(fmt.Sprintf(format, v...))
}

func err(v ...any) {
	v = append([]any{"\033[1;31m"}, v...)
	v = append(v, "\033[0m")
	loggerError.Println(v...)
}

func errf(format string, v ...any) {
	err(fmt.Sprintf(format, v...))
}

var spinners = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

func GetSpinner() func() string {
	var index int = 0
	return func() string {
		index = (index + 1) % len(spinners)
		return spinners[index]
	}
}
