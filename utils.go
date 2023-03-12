package gone

import "time"

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
