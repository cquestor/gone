package gone

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// spin 生成加载图标
var spin = Spinner()

// Debounce 防抖
func Debounce(after time.Duration) func(func()) {
	var timer *time.Timer
	return func(f func()) {
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(after, f)
	}
}

// Loading 加载动画
func Loading(done chan int, message string) {
	for {
		select {
		case <-done:
			return
		default:
			fmt.Printf("\r\033[1;32m%s %s \033[0m", spin(), message)
			time.Sleep(time.Millisecond * 100)
		}
	}
}

// ClearTerm 清屏
func ClearTerm() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// 输出banner
func banner() {
	fmt.Println(" \033[1;32m   ______   \033[1;36m____     \033[1;33m_   __    \033[1;31m______\033[0m")
	fmt.Println(" \033[1;32m  / ____/  \033[1;36m/ __ \\   \033[1;33m/ | / /   \033[1;31m/ ____/\033[0m")
	fmt.Println(" \033[1;32m / / __   \033[1;36m/ / / /  \033[1;33m/  |/ /   \033[1;31m/ __/   \033[0m")
	fmt.Println(" \033[1;32m/ /_/ /  \033[1;36m/ /_/ /  \033[1;33m/ /|  /   \033[1;31m/ /___   \033[0m")
	fmt.Println(" \033[1;32m\\____/   \033[1;36m\\____/  \033[1;33m/_/ |_/   \033[1;31m/_____/   \033[0m")
	fmt.Println()
}
