package elevator

import (
	"net"
	"sync"
	"consts"
	"time"
	"log"
	"encoding/json"
	"network"
	"container/list"
)

func getClientIPAddr() (string) {
	return "localhost"
}



// Master
type Master struct {
	queue   *consts.Queue
	mux     sync.Mutex
	slaveDB *SlavesDB
}

func (m *Master) GetQueue() *consts.Queue {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.queue
}

func (m *Master) GetDB() *SlavesDB {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.slaveDB
}

//func (m *Master) sendToSlave(ip string, order consts.ButtonEvent) { //TODO uncomment when ready
func (m *Master) sendToSlave(notification consts.NotificationData) {
	conn := network.GetSlaveTestSendConn() // TODO get conn from somewhere

	data := GetNotification(notification)
	conn.Write(data)
}

func (m *Master) broadcastToSlaves(notification consts.NotificationData) {
	conn := network.GetSlaveTestSendConn() // TODO user actual broadcast function

	data := GetNotification(notification)
	//log.Println(consts.White, "broadcast", notification)

	conn.Write(data)
}

func (m *Master) masterOrderHandler() {
	for {
		dbData := m.GetQueue().Peek()
		if dbData != nil {
			order := dbData.(consts.ButtonEvent)
			//log.Println(consts.White, "peek of db", order)

			elData := m.GetDB().findFreeElevator(order.Floor)
			if elData != nil {

				// force this elevator to busy (don't wait for periodic update)
				// ignore next 5 updates from specific slave
				item := elData.(dbItem)
				m.GetDB().update(dbItem{
					ip: item.ip,
					ignore: 5,
					data: consts.PeriodicData{
						Floor: item.data.Floor,
						Direction: item.data.Direction,
						CabQueue: item.data.CabQueue,
						Ready: false,
					},
				})

				//ip := elData.(string) //TODO return non string type - net.Conn maybe?
				m.GetQueue().Pop()
				log.Println(consts.White, "parsed order", order, consts.Neutral)

				notification := consts.NotificationData{
					Code: consts.MasterHallOrder,
					Data: GetRawJSON(order),
				}

				m.sendToSlave(notification)
			} else {
				//log.Println(consts.White, "no free elevator")
			}
		}
		
		
		
		time.Sleep(pollRate)
	}
}

func (m *Master) listenIncomingMsg(conn *net.UDPConn) {
	var typeJson consts.NotificationData
	buffer := make([]byte, 8192)

	for {
		n, err := conn.Read(buffer[0:])
		if err != nil {
			log.Println(consts.White, "reading master failed")
			log.Fatal(err)
		}
		/*log.Println(consts.White, string(buffer))*/
		//log.Println(consts.White, buffer)
		if len(buffer) > 0 {
			//log.Println(consts.White, string(buffer))
			err2 := json.Unmarshal(buffer[0:n], &typeJson)
			if err2 != nil {
				log.Println(consts.White, "unmarshal master failed")
				log.Fatal(err2)
			}

			//log.Println(consts.White, "received", typeJson)


			switch typeJson.Code {
			case consts.SlavePeriodicMsg:
				data := consts.PeriodicData{}
				json.Unmarshal(typeJson.Data, &data)

				//log.Println(consts.White, "-------------------------------")
				//log.Println(consts.White, "<- periodic")
				//log.Println(consts.White, "floor:", typeJson.Floor)
				//log.Println(consts.White, "direction:", typeJson.Direction)
				//log.Println(consts.White, "-------------------------------")
				ip := getClientIPAddr()
				m.GetDB().storeData(ip, data)

			case consts.SlaveHallOrder:
				order := consts.ButtonEvent{}
				json.Unmarshal(typeJson.Data, &order)
				log.Println(consts.White, "<- hall order", consts.Neutral)
				//log.Println(consts.White, "-------------------------------")
				//log.Println(consts.White, "received order:", order)
				//log.Println(consts.White, "-------------------------------")
				m.GetQueue().Push(order)

				// TODO broadcast all slaves to turn on light bulbs
				notification := consts.NotificationData{
					Code: consts.MasterHallLight,
					Data: GetRawJSON(order),
				}
				m.broadcastToSlaves(notification)
			}
		}

	}
}

/**
 * defer old instance
 * create Master
 * listen for incoming notifications and orders
 * store notifications to DB
 * handle orders and store to queue
 * broadcast light indicators
 * sync DB with Backup
 * ping all slaves/backup
 * do same things as Slave
 */
func StartMaster(masterConn *net.UDPConn, listenConn *net.UDPConn) {
	ElevatorState.SetMasterConn(masterConn) // master conn has to be available for all roles

	slavesDB := SlavesDB{}
	master := Master{
		consts.NewQueue(),
		sync.Mutex{},
		&slavesDB,
	}

	go master.listenIncomingMsg(listenConn)
	go master.masterOrderHandler()
}



// Database of slaves
type dbItem struct {
	ip 			string
	ignore		int
	data 		consts.PeriodicData
}

type SlavesDB struct {
	list list.List		// list of dbItems
	mux  sync.Mutex
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

			//v.data = item.data
			//v.data.CabQueue = queue
		}
	}
}

func (i *SlavesDB) storeData(ip string, data consts.PeriodicData)  {
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

func (i *SlavesDB) findFreeElevator(floor int) interface{} {
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
