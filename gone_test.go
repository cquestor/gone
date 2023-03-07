package gone_test

import (
	"fmt"
	"os"
	"testing"

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
