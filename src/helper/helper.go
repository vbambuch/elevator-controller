package helper

import (
	"consts"
)

type ASCFloors []consts.ButtonEvent
type DESCFloors []consts.ButtonEvent

func (a ASCFloors) Len() int           { return len(a) }
func (a ASCFloors) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ASCFloors) Less(i, j int) bool { return a[i].Floor < a[j].Floor }


func (a DESCFloors) Len() int           { return len(a) }
func (a DESCFloors) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a DESCFloors) Less(i, j int) bool { return a[i].Floor > a[j].Floor }


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

