package helper

import (
	"time"
	"consts"
)

func Timeout(ms time.Duration, timeout chan<- bool) {
	time.Sleep(ms * time.Millisecond)
	timeout <- true
}

func QueueToArray(queue consts.Queue) ([]consts.ButtonEvent) {
	var result []consts.ButtonEvent
	for {
		if queue.Count != 0 {
			item := queue.Pop().(consts.ButtonEvent)
			result = append(result, item)
		} else {
			break
		}
	}
	return result
}
