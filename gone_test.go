package gone_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cquestor/gone"
)

func TestConfig(t *testing.T) {
	config := gone.AppConfig{}
	if err := config.ParseFile(gone.DEFAULT_CONFIG_PATH); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("配置文件不存在，使用默认配置")
		} else {
			t.Fatal(err)
		}
	}
	fmt.Println(config)
}

func TestLogger(t *testing.T) {
	gone.LogInfo("信息日志")
	gone.LogInfof("%s信息输出", "格式化")
	gone.LogWarn("警告日志")
	gone.LogWarnf("%s警告输出", "格式化")
	gone.LogError("错误日志")
	gone.LogErrorf("%s错误输出", "格式化")
	spinner := gone.GetSpinner()
	for i := 0; i < 30; i++ {
		fmt.Print("\r" + spinner() + " Rebuilding... ")
		time.Sleep(time.Millisecond * 100)
	}
	fmt.Println("\r" + " √" + " Build Success!")
}

func TestWatch(t *testing.T) {
	watcher, err := gone.NewWatcher()
	if err != nil {
		t.Fatal(err.Error())
	}
	watcher.AddWatch(".")
	watcher.Watch()
}

func TestRun(t *testing.T) {
	g := gone.New()
	g.Run()
}
