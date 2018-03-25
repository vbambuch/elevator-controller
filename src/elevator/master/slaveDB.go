package master

import (
	"consts"
	"container/list"
	"sync"
	"log"
)

// Database of slaves
type SlaveData struct {
	Floor     int
	Direction consts.MotorDirection
	CabQueue  consts.Queue
	Ready 	  bool
}

type dbItem struct {
	ip     string
	ignore int
	data   SlaveData
}

type SlavesDB struct {
	list list.List		// list of dbItems
	mux  sync.Mutex
}

func (i *SlavesDB) dump() {
	i.mux.Lock()
	defer i.mux.Unlock()

	for e := i.list.Front(); e != nil; e = e.Next() {
		data := e.Value.(dbItem).data
		log.Println(consts.Yellow, "floor:", data.Floor, consts.Neutral)
		log.Println(consts.Yellow, "direction:", data.Direction, consts.Neutral)
		log.Println(consts.Yellow, "queue:", data.CabQueue, consts.Neutral)
		log.Println(consts.Yellow, "ready:", data.Ready, consts.Neutral)
	}
}

func (i *SlavesDB) exists(ip string) (bool) {
	i.mux.Lock()
	defer i.mux.Unlock()

	for e := i.list.Front(); e != nil; e = e.Next() {
		if e.Value.(dbItem).ip == ip {
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
		if v.ip == item.ip {
			//queue := v.data.CabQueue // keep previous queue
			//log.Println(consts.White, "db item", v.ignore)
			//log.Printf("db item %+v", v.data.CabQueue.Len())

			if v.ignore > 0 {
				e.Value = dbItem{
					ip: v.ip,
					ignore: v.ignore -1,
					data: v.data,
				}
			} else {
				e.Value = dbItem{
					ip: v.ip,
					ignore: item.ignore,
					data: item.data,
				}
			}
		}
	}
}

func (i *SlavesDB) storeData(ip string, data SlaveData)  {
	item := dbItem{ip, 0,data}
	if i.exists(ip) {
		i.update(item)
		//log.Println(consts.White, "db item", item.data.Ready)
	} else {
		i.list.PushBack(item)
	}

	//log.Println(consts.White, "db", i)
}

func (i *SlavesDB) findElevatorsOnFloor(floor int) interface{} {
	i.mux.Lock()
	defer i.mux.Unlock()

	for e := i.list.Front(); e != nil; e = e.Next() {

		elData := e.Value.(dbItem).data
		if elData.Floor == floor {
			return elData
		}
	}
	return nil
}

func (i *SlavesDB) findFreeElevator() interface{} {
	i.mux.Lock()
	defer i.mux.Unlock()

	for e := i.list.Front(); e != nil; e = e.Next() {
		item := e.Value.(dbItem)
		//log.Println(consts.White, "db item", item)

		if item.data.Ready {
			return item
		}
	}
	return nil
}

