package helper

import (
	"consts"
	"sort"
)

type ASCFloors []consts.ButtonEvent
type DESCFloors []consts.ButtonEvent
type ByQueue []consts.DBItem
type ByFloorDiff []consts.FreeElevatorItem


func (a ASCFloors) Len() int           { return len(a) }
func (a ASCFloors) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ASCFloors) Less(i, j int) bool { return a[i].Floor < a[j].Floor }

func (a DESCFloors) Len() int           { return len(a) }
func (a DESCFloors) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a DESCFloors) Less(i, j int) bool { return a[i].Floor > a[j].Floor }

func (a ByQueue) Len() int           { return len(a) }
func (a ByQueue) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByQueue) Less(i, j int) bool { return len(a[i].Data.CabArray) < len(a[j].Data.CabArray) }

func (a ByFloorDiff) Len() int           { return len(a) }
func (a ByFloorDiff) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByFloorDiff) Less(i, j int) bool { return a[i].FloorDiff < a[j].FloorDiff }



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

func GetShortestQueueElevator(suitableArray []consts.DBItem) interface{} {
	if len(suitableArray) > 0 {
		sort.Sort(ByQueue(suitableArray))
		return suitableArray[0]
	}
	return nil
}

func GetLowestDiffElevator(freeArray []consts.FreeElevatorItem) interface{} {
	if len(freeArray) > 0 {
		sort.Sort(ByFloorDiff(freeArray))
		return freeArray[0].Data
	}
	return nil
}
