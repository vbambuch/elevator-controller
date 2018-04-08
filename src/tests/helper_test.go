package tests

import (
	"testing"
	"consts"
	"helper"
	"sort"
	"log"
)

func TestQueueToList(t *testing.T) {
	queue := consts.NewQueue()

	order1 := consts.ButtonEvent{Floor: 1, Button: 2}
	order2 := consts.ButtonEvent{Floor: 2, Button: 2}
	order3 := consts.ButtonEvent{Floor: 3, Button: 1}

	array := []consts.ButtonEvent{order1, order2, order3}

	queue.Push(order1)
	queue.Push(order2)
	queue.Push(order3)

	result := helper.QueueToArray(*queue)
	for i, v := range result {
		if v != array[i] {
			t.Errorf("%+d doesn't match %+d", v, array[i])
		}
	}
}

func TestSorting(t *testing.T) {

	order1 := consts.ButtonEvent{Floor: 1, Button: 2}
	order2 := consts.ButtonEvent{Floor: 0, Button: 2}
	order3 := consts.ButtonEvent{Floor: 3, Button: 2}

	array := []consts.ButtonEvent{order1, order2, order3}
	ascArray := []int{0, 1, 3}
	descArray := []int{3, 1, 0}

	sort.Sort(helper.ASCFloors(array))
	log.Println(array)
	for i, v := range array {
		if v.Floor != ascArray[i] {
			t.Errorf("%+d doesn't match %+d", v, array[i])
		}
	}

	sort.Sort(helper.DESCFloors(array))
	log.Println(array)
	for i, v := range array {
		if v.Floor != descArray[i] {
			t.Errorf("%+d doesn't match %+d", v, array[i])
		}
	}
}
