package gone_test

import (
	"os"
	"testing"

	"github.com/cquestor/gone"
)

func TestWatch(t *testing.T) {
	path, _ := os.Getwd()
	if w, err := gone.NewWatcher(path, []string{".gone"}, []string{"./utils"}); err != nil {
		t.Fatal(err)
	} else {
		w.Start()
	}
}
