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
	"helper"
)


// Master
type Master struct {
	orderList 	*list.List
	mux       	sync.Mutex
	slaveDB   	*SlavesDB
	backupConn	*net.UDPConn
}

func (m *Master) getOrderList() *list.List {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.orderList
}

func (m *Master) getDB() *SlavesDB {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.slaveDB
}

func (m *Master) setBackupConn(conn *net.UDPConn) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.backupConn = conn
}

func (m *Master) getBackupConn() *net.UDPConn {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.backupConn
}

func (m *Master) newOrder(order consts.ButtonEvent) bool {
	m.mux.Lock()
	defer m.mux.Unlock()
	for el := m.orderList.Front(); el != nil; el = el.Next() {
		data := el.Value.(consts.HallOrders).Order
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
		data := el.Value.(consts.HallOrders).Order
		if data.Floor == order.Floor && data.Button == order.Button {
			m.orderList.Remove(el)
		}
	}
}

func (m *Master) getOrder() (interface{}) {
	m.mux.Lock()
	defer m.mux.Unlock()
	for el := m.orderList.Front(); el != nil; el = el.Next() {
		data := el.Value.(consts.HallOrders)
		if data.AssignedTo == consts.Unassigned {
			return data.Order
		}
	}
	return nil
}

func (m *Master) assignedToSlave(order consts.ButtonEvent, ipAddr string) {
	m.mux.Lock()
	defer m.mux.Unlock()
	for el := m.orderList.Front(); el != nil; el = el.Next() {
		data := el.Value.(consts.HallOrders).Order
		if data.Floor == order.Floor && data.Button == order.Button {
			el.Value = consts.HallOrders{
				Order:      order,
				AssignedTo: ipAddr,
			}
		}
	}
}

func (m *Master) makeHallOrderAvailable(ipAddr string)  {
	m.mux.Lock()
	defer m.mux.Unlock()
	for el := m.orderList.Front(); el != nil; el = el.Next() {
		data := el.Value.(consts.HallOrders)
		if data.AssignedTo == ipAddr {
			el.Value = consts.HallOrders{
				Order:      data.Order,
				AssignedTo: consts.Unassigned,
			}
			log.Println(consts.Yellow, "Available again:", data.Order, consts.Neutral)
		}
	}
}

func (m *Master) updateHallButtons(conn *net.UDPConn)  {
	m.mux.Lock()
	orderList := m.orderList.Front()
	m.mux.Unlock()
	for el := orderList; el != nil; el = el.Next() {
		order := el.Value.(consts.HallOrders).Order
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
		data := el.Value.(consts.HallOrders)
		log.Print(consts.Yellow,
			" Floor:", data.Order.Floor,
			" Button:", data.Order.Button,
			" Assigned to: ", data.AssignedTo,
		consts.Neutral)
	}
	log.Println()
	log.Println(consts.Yellow, "-----", consts.Neutral)
}

func (m *Master) sendToSlave(conn *net.UDPConn, notification interface{}) {
	data := common.GetNotification(notification)
	if conn != nil {
		_, err := conn.Write(data)
		helper.HandleError(err, "Sending to slave")
		//log.Println(consts.White, "Sent to:", conn.RemoteAddr(), consts.Neutral)
		//log.Println(consts.White, "Sent", n, "bytes", consts.Neutral)

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
				// skip this order meanwhile
				m.assignedToSlave(order, ip)
				//m.dumpList()

				log.Println(consts.White, ip, ": parsed order", order, consts.Neutral)
				//m.getOrderList().Dump()

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
	hallList := m.getOrderList()

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
					hallOrder := consts.HallOrders{
						Order:      order,
						AssignedTo: consts.Unassigned,
					}
					hallList.PushBack(hallOrder)

					//m.getOrderList().Dump()

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

func (m *Master) synchronizeDataWithBackup() {
	conn := network.GetSendConn(network.GetBroadcastAddress()+consts.BackupPort)

	for {
		backupSync := consts.BackupSync{
			SlavesList: helper.ListToSlavesArray(*m.getDB().getList()),
			OrderList:  helper.ListToOrderArray(*m.getOrderList()),
			Timestamp:  time.Now(),
		}
		m.sendToSlave(conn, backupSync)
		time.Sleep(100 * consts.PollRate)
	}
}

func (m *Master) recreateSlavesConnections() {
	for el := m.getDB().getList().Front(); el != nil; el = el.Next() {
		data := el.Value.(consts.DBItem)
		log.Println(consts.White, "New connection for:", data.Data.ListenIP, consts.Neutral)
		conn := network.GetSendConn(data.Data.ListenIP)
		el.Value = consts.DBItem{
			ClientConn: conn,
			Timestamp: 	data.Timestamp,
			Data: 		data.Data,
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
func StartMaster(backupData consts.BackupSync) {
	listenConn := network.GetListenConn(network.GetBroadcastAddress()+consts.MasterPort)

	orderList := helper.OrderArrayToList(backupData.OrderList)
	slavesList := helper.SlaveArrayToList(backupData.SlavesList)

	slavesDB := SlavesDB{list: slavesList}
	master := Master{
		orderList,
		sync.Mutex{},
		&slavesDB,
		nil,
	}

	if slavesList.Len() != 0 { 	// recovered from backup
		log.Println(consts.White, "Recovering master from backup...", consts.Neutral)
		master.recreateSlavesConnections()
	} else {					// brand new master
		log.Println(consts.White, "Starting new master...", consts.Neutral)
	}


	go master.listenIncomingMsg(listenConn)
	go master.masterHallOrderHandler()
	go master.synchronizeDataWithBackup()
}






