package tests

import (
	"testing"
	"consts"
	"helper"
)

func TestQueueToList(t *testing.T)  {
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
