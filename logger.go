package gone

import (
	"log"
	"os"
)

var (
	loggerInfo  = log.New(os.Stderr, "[INFO]", log.LstdFlags)
	loggerWarn  = log.New(os.Stderr, "[WARN]", log.LstdFlags)
	loggerError = log.New(os.Stderr, "[ERROR]", log.LstdFlags)
)

var (
	LogInfo   = loggerInfo.Println
	LogInfof  = loggerInfo.Printf
	LogWarn   = loggerWarn.Println
	LogWarnf  = loggerWarn.Printf
	LogError  = loggerError.Println
	LogErrorf = loggerError.Printf
)
