package master

import (
	"consts"
	"container/list"
	"sync"
	"log"
	"sort"
	"net"
	"network"
)

// Sorting dbItems
type ByQueue []dbItem

func (a ByQueue) Len() int           { return len(a) }
func (a ByQueue) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByQueue) Less(i, j int) bool { return len(a[i].data.CabArray) < len(a[j].data.CabArray) }


// Database of slaves
//type SlaveData struct {
//	Floor      int
//	Direction  consts.MotorDirection
//	OrderArray []consts.ButtonEvent
//	Free      bool
//}

type dbItem struct {
	clientConn 		*net.UDPConn
	ignore 			int		//ignore number of incoming messages
	data   			consts.PeriodicData
}

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
		data := e.Value.(dbItem).data
		log.Println(consts.Yellow, "----------", consts.Neutral)
		log.Println(consts.Yellow, "ip:", data.ListenIP, consts.Neutral)
		log.Println(consts.Yellow, "floor:", data.Floor, consts.Neutral)
		log.Println(consts.Yellow, "direction:", data.Direction, consts.Neutral)
		log.Println(consts.Yellow, "queue:", data.CabArray, consts.Neutral)
		log.Println(consts.Yellow, "ready:", data.Free, consts.Neutral)
		log.Println(consts.Yellow, "processing:", data.HallProcessing, consts.Neutral)
		log.Println(consts.Yellow, "-----", consts.Neutral)
	}
}

func (i *SlavesDB) exists(ip string) (bool) {
	i.mux.Lock()
	defer i.mux.Unlock()

	for e := i.list.Front(); e != nil; e = e.Next() {
		el := e.Value.(dbItem)
		if el.data.ListenIP == ip {
			return true
		}
	}
	return false
}

func (i *SlavesDB) update(item dbItem) {
	i.mux.Lock()
	defer i.mux.Unlock()

	for e := i.list.Front(); e != nil; e = e.Next() {
		v := e.Value.(dbItem)
		if v.data.ListenIP == item.data.ListenIP {
			//queue := v.data.OrderArray // keep previous queue
			//log.Println(consts.White, "db item", v.ignore, v.data.Free)
			//log.Printf("db item %+v", v.data.OrderArray.Len())

			if v.ignore > 0 {
				e.Value = dbItem{
					clientConn: v.clientConn,	// keep previous conn
					ignore: v.ignore -1,
					data: v.data,
				}
			} else {
				e.Value = dbItem{
					clientConn: v.clientConn,	// keep previous conn
					ignore: item.ignore,
					data: item.data,
				}
			}
			return
		}
	}
	//log.Println(consts.Yellow, "trying to push", consts.Neutral)

	clientConn := network.GetSlaveSendConn(item.data.ListenIP)
	item.clientConn = clientConn
	i.list.PushBack(item)
	log.Println(consts.White, "ListenIP:", item.data.ListenIP, consts.Neutral)
	//log.Println(consts.White, "Conn:", item.clientConn.RemoteAddr(), consts.Neutral)
}

func (i *SlavesDB) storeData(data consts.PeriodicData)  {
	i.mux.Lock()
	defer i.mux.Unlock()
	item := dbItem{nil,0,data}
	i.update(item)
}

func (i *SlavesDB) findElevatorOnFloor(floor int) interface{} {
	i.mux.Lock()
	defer i.mux.Unlock()

	var onFloorArray []dbItem
	for e := i.list.Front(); e != nil; e = e.Next() {
		elItem := e.Value.(dbItem)
		if elItem.data.Floor == floor && elItem.data.Free {
			onFloorArray = append(onFloorArray, elItem)
		}
	}

	// get elevator with the shortest cab queue
	return getShortestQueueElevator(onFloorArray)
}

func (i *SlavesDB) findFreeElevator() interface{} {
	i.mux.Lock()
	defer i.mux.Unlock()

	for e := i.list.Front(); e != nil; e = e.Next() {
		item := e.Value.(dbItem)
		//log.Println(consts.White, "db item", item)

		if item.data.Free {
			return item
		}
	}
	return nil
}

func (i *SlavesDB) findSameDirection(order consts.ButtonEvent) interface{} {
	i.mux.Lock()
	defer i.mux.Unlock()

	var suitableArray []dbItem
	for e := i.list.Front(); e != nil; e = e.Next() {
		item := e.Value.(dbItem)
		//log.Println(consts.White, "db item", item)

		if !item.data.HallProcessing && suitableElevator(item.data.CabArray, item.data.Floor, order) {
			suitableArray = append(suitableArray, item)
		} else {
		//log.Println(consts.Yellow, "not suitable", item.data.CabArray, consts.Neutral)
		}
	}

	// get elevator with the shortest cab queue
	return getShortestQueueElevator(suitableArray)
}

/**
 * Main function for elevator searching
 */
func (i *SlavesDB) findElevator(order consts.ButtonEvent) interface{} {
	message := "on floor"
	elevator := i.findElevatorOnFloor(order.Floor)
	if elevator == nil {
		message = "free"
		elevator = i.findFreeElevator()
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

func getShortestQueueElevator(suitableArray []dbItem) interface{} {
	if len(suitableArray) > 0 {
		sort.Sort(ByQueue(suitableArray))
		return suitableArray[0]
	}
	return nil
}
