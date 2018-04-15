//
//  queue.go
//
//  Created by Hicham Bouabdallah
//  Copyright (c) 2012 SimpleRocket LLC
//
//  Permission is hereby granted, free of charge, to any person
//  obtaining a copy of this software and associated documentation
//  files (the "Software"), to deal in the Software without
//  restriction, including without limitation the rights to use,
//  copy, modify, merge, publish, distribute, sublicense, and/or sell
//  copies of the Software, and to permit persons to whom the
//  Software is furnished to do so, subject to the following
//  conditions:
//
//  The above copyright notice and this permission notice shall be
//  included in all copies or substantial portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
//  EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
//  OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//  NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
//  HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
//  FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
//  OTHER DEALINGS IN THE SOFTWARE.
//

package consts

import (
	"sync"
)

type queuenode struct {
	data interface{}
	next *queuenode
}

//	A go-routine safe FIFO (first in first out) data stucture.
type Queue struct {
	Head  *queuenode
	Tail  *queuenode
	Count int
	Lock  *sync.Mutex
}

//	Creates a new pointer to a new queue.
func NewQueue() *Queue {
	q := &Queue{}
	q.Lock = &sync.Mutex{}
	return q
}

//	Returns the number of elements in the queue (i.e. size/length)
//	go-routine safe.
func (q *Queue) Len() int {
	q.Lock.Lock()
	defer q.Lock.Unlock()
	return q.Count
}

//	Pushes/inserts a value at the end/Tail of the queue.
//	Note: this function does mutate the queue.
//	go-routine safe.
func (q *Queue) Push(item interface{}) {
	q.Lock.Lock()
	defer q.Lock.Unlock()

	n := &queuenode{data: item}

	if q.Tail == nil {
		q.Tail = n
		q.Head = n
	} else {
		q.Tail.next = n
		q.Tail = n
	}
	q.Count++
}

//	Returns the value at the front of the queue.
//	i.e. the oldest value in the queue.
//	Note: this function does mutate the queue.
//	go-routine safe.
func (q *Queue) Pop() interface{} {
	q.Lock.Lock()
	defer q.Lock.Unlock()

	if q.Head == nil {
		return nil
	}

	n := q.Head
	q.Head = n.next

	if q.Head == nil {
		q.Tail = nil
	}
	q.Count--

	return n.data
}

//	Returns a read value at the front of the queue.
//	i.e. the oldest value in the queue.
//	Note: this function does NOT mutate the queue.
//	go-routine safe.
func (q *Queue) Peek() interface{} {
	q.Lock.Lock()
	defer q.Lock.Unlock()

	n := q.Head
	if n == nil {
		return nil
	}

	return n.data
}

// Check if element exists in queue
func (q *Queue) NewOrder(order ButtonEvent) bool {
	q.Lock.Lock()
	tmp := *q
	q.Lock.Unlock()
	for el := tmp.Pop(); el != nil; el = tmp.Pop() {
		data := el.(ButtonEvent)
		if data.Floor == order.Floor && data.Button == order.Button {
			//log.Println(Green, "Order:", data, Neutral)
			return false
		}
	}
	//log.Println(Green, "Is new:", order, Neutral)

	return true
}
