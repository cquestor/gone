package gone_test

import (
	"fmt"
	"testing"

	"github.com/cquestor/gone"
)

func TestWatch(t *testing.T) {
	w, err := gone.NewWatcher("./", nil, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	sigChan := make(chan int, 1)
	w.Start(sigChan)
	for {
		<-sigChan
		fmt.Println("热更新!")
	}
}

func TestLog(t *testing.T) {
	gone.LogInfo("信息输出")
	gone.LogInfof("信息%s输出\n", "格式化")
	gone.LogWarn("警告输出")
	gone.LogWarnf("警告%s输出\n", "格式化")
	gone.LogErr("错误输出")
	gone.LogErrf("错误%s输出\n", "格式化")
}
