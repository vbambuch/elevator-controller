package helper

import "time"

func Timeout(ms time.Duration, timeout chan<- bool) {
	time.Sleep(ms * time.Millisecond)
	timeout <- true
}
