package master

import (
	"net"
	"sync"
	"consts"
	"time"
	"log"
	"encoding/json"
	"elevator/common"
	"network"
	"container/list"
)


// Master
type Master struct {
	orderList *list.List
	mux       sync.Mutex
	slaveDB   *SlavesDB
}

type hallOrders struct {
	order      consts.ButtonEvent
	assignedTo string
}

func (m *Master) getList() *list.List {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.orderList
}

func (m *Master) getDB() *SlavesDB {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.slaveDB
}

func (m *Master) newOrder(order consts.ButtonEvent) bool {
	m.mux.Lock()
	defer m.mux.Unlock()
	for el := m.orderList.Front(); el != nil; el = el.Next() {
		data := el.Value.(hallOrders).order
		if data.Floor == order.Floor && data.Button == order.Button {
			return false
		}
	}
	return true
}

func (m *Master) deleteOrder(order consts.ButtonEvent)  {
	m.mux.Lock()
	defer m.mux.Unlock()
	for el := m.orderList.Front(); el != nil; el = el.Next() {
		data := el.Value.(hallOrders).order
		if data.Floor == order.Floor && data.Button == order.Button {
			m.orderList.Remove(el)
		}
	}
}

func (m *Master) getOrder() (interface{}) {
	m.mux.Lock()
	defer m.mux.Unlock()
	for el := m.orderList.Front(); el != nil; el = el.Next() {
		data := el.Value.(hallOrders)
		if data.assignedTo == consts.Unassigned {
			return data.order
		}
	}
	return nil
}

func (m *Master) assignedToSlave(order consts.ButtonEvent, ipAddr string) {
	m.mux.Lock()
	defer m.mux.Unlock()
	for el := m.orderList.Front(); el != nil; el = el.Next() {
		data := el.Value.(hallOrders).order
		if data.Floor == order.Floor && data.Button == order.Button {
			el.Value = hallOrders{
				order:      order,
				assignedTo: ipAddr,
			}
		}
	}
}

func (m *Master) makeHallOrderAvailable(ipAddr string)  {
	m.mux.Lock()
	defer m.mux.Unlock()
	for el := m.orderList.Front(); el != nil; el = el.Next() {
		data := el.Value.(hallOrders)
		if data.assignedTo == ipAddr {
			el.Value = hallOrders{
				order:      data.order,
				assignedTo: consts.Unassigned,
			}
			log.Println(consts.Yellow, "Available again:", data.order, consts.Neutral)
		}
	}
}

func (m *Master) updateHallButtons(conn *net.UDPConn)  {
	m.mux.Lock()
	orderList := m.orderList.Front()
	m.mux.Unlock()
	for el := orderList; el != nil; el = el.Next() {
		order := el.Value.(hallOrders).order
		orderData := consts.NotificationData{
			Code: consts.MasterHallLight,
			Data: common.GetRawJSON(order),
		}

		m.sendToSlave(conn, orderData)
	}
}

func (m *Master) dumpList() {
	m.mux.Lock()
	defer m.mux.Unlock()
	log.Println(consts.Yellow, "----------", consts.Neutral)
	for el := m.orderList.Front(); el != nil; el = el.Next() {
		data := el.Value.(hallOrders)
		log.Print(consts.Yellow,
			" Floor:", data.order.Floor,
			" Button:", data.order.Button,
			" Assigned to: ", data.assignedTo,
		consts.Neutral)
	}
	log.Println()
	log.Println(consts.Yellow, "-----", consts.Neutral)
}

func (m *Master) sendToSlave(conn *net.UDPConn, notification consts.NotificationData) {
	data := common.GetNotification(notification)
	if conn != nil {
		//log.Println(consts.White, "Send to:", conn.RemoteAddr(), consts.Neutral)
		conn.Write(data)
	}
}

func (m *Master) broadcastToSlaves(data consts.NotificationData) {
	//log.Println(consts.White, "broadcast", n)
	slaves := m.getDB().getList()
	for el := slaves.Front(); el != nil; el = el.Next() {
		conn := el.Value.(consts.DBItem).ClientConn
		m.sendToSlave(conn, data)
	}
}

func (m *Master) masterHallOrderHandler() {
	db := m.getDB()

	for {
		// filter out outdated slaves from DB
		outdated := db.deleteOutdatedSlaves()
		if outdated != consts.NoOutdated {
			// make hallOrder available again
			m.makeHallOrderAvailable(outdated)
		}

		queue := m.getOrder()
		if queue != nil { // empty orderList test
			order := queue.(consts.ButtonEvent)
			//log.Println(consts.White, "peek of db", order)



			elData := db.findElevator(order)
			if elData != nil {
				//log.Println(consts.Red, "elData:", elData, consts.Neutral)

				// force this elevator to busy (don't wait for periodic update)
				// ignore next 10 updates from specific slave
				item := elData.(consts.DBItem)
				ip := item.Data.ListenIP

				// ignore return value for this case
				db.updateOrInsert(consts.DBItem{
					ClientConn: 	item.ClientConn,
					Ignore: 		10,
					Timestamp: 		item.Timestamp,
					Data: consts.PeriodicData{
						ListenIP:       ip,
						Floor:          item.Data.Floor,
						Direction:      item.Data.Direction,
						OrderArray:     item.Data.OrderArray,
						Free:           false,
						HallProcessing: true,
						Stopped:		item.Data.Stopped,
					},
				})

				// order is in progress => wait for resolving
				// skip whit order meanwhile
				m.assignedToSlave(order, ip)
				//m.dumpList()

				log.Println(consts.White, ip, ": parsed order", order, consts.Neutral)
				//m.getList().Dump()

				orderData := consts.NotificationData{
					Code: consts.MasterHallOrder,
					Data: common.GetRawJSON(order),
				}

				m.sendToSlave(item.ClientConn, orderData)
			} else {
				//log.Println(consts.White, "no free elevator")
			}
		}
		time.Sleep(10 * consts.PollRate)
	}
}

func (m *Master) listenIncomingMsg(conn *net.UDPConn) {
	var typeJson consts.NotificationData
	buffer := make([]byte, 8192)
	hallList := m.getList()

	for {
		n, err := conn.Read(buffer[0:])
		if err != nil {
			log.Println(consts.White, "reading master failed")
			log.Fatal(err)
		}
		//log.Println(consts.White, string(buffer))
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
				data := consts.PeriodicData{} // for parsing of incoming message
				json.Unmarshal(typeJson.Data, &data)

				conn := m.getDB().storeData(data)
				// new elevator has been connected => update its hall buttons
				if conn != nil {
					m.updateHallButtons(conn)
				}

				//m.getDB().dump()
			case consts.SlaveHallOrder:
				order := consts.ButtonEvent{}
				json.Unmarshal(typeJson.Data, &order)
				log.Println(consts.White, "<- hall order", consts.Neutral)
				//log.Println(consts.White, "orderList length:", hallList.Len(), consts.Neutral)

				if m.newOrder(order) {
					hallOrder := hallOrders{
						order:      order,
						assignedTo: consts.Unassigned,
					}
					hallList.PushBack(hallOrder)

					//m.getList().Dump()

					//broadcast all slaves to turn on light bulbs
					notification := consts.NotificationData{
						Code: consts.MasterHallLight,
						Data: common.GetRawJSON(order),
					}
					m.broadcastToSlaves(notification)
				}

			case consts.ClearHallOrder:
				order := consts.ButtonEvent{}
				json.Unmarshal(typeJson.Data, &order)

				// delete from hallList
				m.deleteOrder(order)

				//broadcast all slaves to turn off light bulbs
				notification := consts.NotificationData{
					Code: consts.ClearHallOrder,
					Data: common.GetRawJSON(order),
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
 * handle orders and store to orderList
 * broadcast light indicators
 * sync DB with Backup
 * ping all slaves/backup
 * do same things as Slave
 */
func StartMaster() {
	listenConn := network.GetListenConn(consts.BListenAddress)

	slavesDB := SlavesDB{}
	master := Master{
		list.New(),
		sync.Mutex{},
		&slavesDB,
	}

	go master.listenIncomingMsg(listenConn)
	go master.masterHallOrderHandler()
}






