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
