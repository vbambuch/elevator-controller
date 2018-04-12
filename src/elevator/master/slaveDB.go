package master

import (
	"consts"
	"container/list"
	"sync"
	"log"
	"network"
	"helper"
	"math"
)



type SlavesDB struct {
	list list.List		// list of dbItems
	mux  sync.Mutex
}

func (i *SlavesDB) getList() (list.List) {
	i.mux.Lock()
	defer i.mux.Unlock()
	return i.list
}

func (i *SlavesDB) dump() {
	i.mux.Lock()
	defer i.mux.Unlock()

	for e := i.list.Front(); e != nil; e = e.Next() {
		data := e.Value.(consts.DBItem).Data
		log.Println(consts.Yellow, "----------", consts.Neutral)
		log.Println(consts.Yellow, "ip:", data.ListenIP, consts.Neutral)
		log.Println(consts.Yellow, "floor:", data.Floor, consts.Neutral)
		log.Println(consts.Yellow, "direction:", data.Direction, consts.Neutral)
		log.Println(consts.Yellow, "queue:", data.OrderArray, consts.Neutral)
		log.Println(consts.Yellow, "ready:", data.Free, consts.Neutral)
		log.Println(consts.Yellow, "processing:", data.HallProcessing, consts.Neutral)
		log.Println(consts.Yellow, "-----", consts.Neutral)
	}
}

func (i *SlavesDB) exists(ip string) (bool) {
	i.mux.Lock()
	defer i.mux.Unlock()

	for e := i.list.Front(); e != nil; e = e.Next() {
		el := e.Value.(consts.DBItem)
		if el.Data.ListenIP == ip {
			return true
		}
	}
	return false
}

func (i *SlavesDB) update(item consts.DBItem) {
	i.mux.Lock()
	defer i.mux.Unlock()

	for e := i.list.Front(); e != nil; e = e.Next() {
		v := e.Value.(consts.DBItem)
		if v.Data.ListenIP == item.Data.ListenIP {
			//queue := v.data.OrderArray // keep previous queue
			//log.Println(consts.White, "db item", v.ignore, v.data.Free)
			//log.Printf("db item %+v", v.data.OrderArray.Len())

			if v.Ignore > 0 {
				e.Value = consts.DBItem{
					ClientConn: v.ClientConn,	// keep previous conn
					Ignore: v.Ignore -1,
					Data: v.Data,
				}
			} else {
				e.Value = consts.DBItem{
					ClientConn: v.ClientConn,	// keep previous conn
					Ignore: item.Ignore,
					Data: item.Data,
				}
			}
			return
		}
	}
	//log.Println(consts.Yellow, "trying to push", consts.Neutral)

	clientConn := network.GetSendConn(item.Data.ListenIP)
	item.ClientConn = clientConn
	i.list.PushBack(item)
	log.Println(consts.White, "ListenIP:", item.Data.ListenIP, consts.Neutral)
	//log.Println(consts.White, "Conn:", item.clientConn.RemoteAddr(), consts.Neutral)
}

func (i *SlavesDB) storeData(data consts.PeriodicData)  {
	item := consts.DBItem{ClientConn: nil, Ignore: 0, Data: data}
	i.update(item)
}



/*
 *	Functions for and assigning of orders to elevators
 */
func suitableElevator(cabArray []consts.ButtonEvent, currFloor int, hallOrder consts.ButtonEvent) bool {
	for _, cabOrder := range cabArray {
		// elevator | hallCall - UP | cabCall => stop on hallCall
		if currFloor < hallOrder.Floor &&
			hallOrder.Floor < cabOrder.Floor &&
			hallOrder.Button == consts.ButtonUP {
			return true
			// cabCall | hallCall - DOWN | elevator => stop on hallCall
		} else if currFloor > hallOrder.Floor &&
			hallOrder.Floor > cabOrder.Floor &&
			hallOrder.Button == consts.ButtonDOWN {
			return true
		}
	}
	return false
}


func (i *SlavesDB) findElevatorOnFloor(floor int) interface{} {
	i.mux.Lock()
	defer i.mux.Unlock()

	var onFloorArray []consts.DBItem
	for e := i.list.Front(); e != nil; e = e.Next() {
		elItem := e.Value.(consts.DBItem)
		if elItem.Data.Floor == floor && elItem.Data.Free && !elItem.Data.Stopped {
			onFloorArray = append(onFloorArray, elItem)
		}
	}

	// get elevator with the shortest cab queue
	return helper.GetShortestQueueElevator(onFloorArray)
}

func (i *SlavesDB) findFreeElevator(floor int) interface{} {
	i.mux.Lock()
	defer i.mux.Unlock()

	var freeArray []consts.FreeElevatorItem
	for e := i.list.Front(); e != nil; e = e.Next() {
		item := e.Value.(consts.DBItem)
		if item.Data.Free && !item.Data.Stopped {
			freeArray = append(freeArray, consts.FreeElevatorItem{
				FloorDiff: math.Abs(float64(item.Data.Floor) - float64(floor)),
				Data: item,
			})
		}
	}

	// get the closest elevator
	return helper.GetLowestDiffElevator(freeArray)
}

func (i *SlavesDB) findSameDirection(order consts.ButtonEvent) interface{} {
	i.mux.Lock()
	defer i.mux.Unlock()

	var suitableArray []consts.DBItem
	for e := i.list.Front(); e != nil; e = e.Next() {
		item := e.Value.(consts.DBItem)
		//log.Println(consts.White, "db item", item)

		if !item.Data.HallProcessing && !item.Data.Stopped &&
			suitableElevator(item.Data.OrderArray, item.Data.Floor, order) {
			suitableArray = append(suitableArray, item)
		} else {
		//log.Println(consts.Yellow, "not suitable", item.data.OrderArray, consts.Neutral)
		}
	}

	// get elevator with the shortest cab queue
	return helper.GetShortestQueueElevator(suitableArray)
}

// Main function for elevator searching
func (i *SlavesDB) findElevator(order consts.ButtonEvent) interface{} {
	message := "on floor"
	elevator := i.findElevatorOnFloor(order.Floor)
	if elevator == nil {
		message = "free"
		elevator = i.findFreeElevator(order.Floor)
		if elevator == nil {
			message = "same direction"
			elevator = i.findSameDirection(order)
			if elevator == nil {
				//TODO
				//log.Println(consts.White, "Elevator not found", consts.Neutral)
			}
		}
	}

	if elevator != nil {
		log.Println(consts.Yellow, "Found: elevator", message, consts.Neutral)
	}

	return elevator
}
