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

/**
 * Reserve specific hall order to some elevator.
 * To be able to assign to different one if previous failed.
 */
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

/**
 * Previous elevator failed => make the order available for others
 */
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

/**
 * New elevator has been connected to the network.
 * Force him to turn on all current hall order lights.
 */
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
	}
}

func (m *Master) broadcastToSlaves(data consts.NotificationData) {
	slaves := m.getDB().getList()
	for el := slaves.Front(); el != nil; el = el.Next() {
		conn := el.Value.(consts.DBItem).ClientConn
		m.sendToSlave(conn, data)
	}
}

/**
 * Get some available hall order and select correct elevator for that order.
 */
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
			elData := db.findElevator(order)
			if elData != nil {
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
						Role:			item.Data.Role,
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

				log.Println(consts.White, ip, ": parsed order", order, consts.Neutral)

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

/**
 * Listen for messages from all elevators on the broadcast address.
 */
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
		if len(buffer) > 0 {
			err2 := json.Unmarshal(buffer[0:n], &typeJson)
			if err2 != nil {
				log.Println(consts.White, "unmarshal master failed")
				log.Fatal(err2)
			}

			switch typeJson.Code {
			case consts.SlavePeriodicMsg:
				data := consts.PeriodicData{} // for parsing of incoming message
				json.Unmarshal(typeJson.Data, &data)

				conn := m.getDB().storeData(data)
				// new elevator has been connected => update its hall buttons
				if conn != nil {
					m.updateHallButtons(conn)
				}
			case consts.SlaveHallOrder:
				order := consts.ButtonEvent{}
				json.Unmarshal(typeJson.Data, &order)
				log.Println(consts.White, "<- hall order", consts.Neutral)

				if m.newOrder(order) {
					hallOrder := consts.HallOrders{
						Order:      order,
						AssignedTo: consts.Unassigned,
					}
					hallList.PushBack(hallOrder)

					//broadcast all slaves to turn on light bulbs
					notification := consts.NotificationData{
						Code: consts.MasterHallLight,
						Data: common.GetRawJSON(order),
					}
					m.broadcastToSlaves(notification)
				}

			case consts.ClearHallOrder:	// elevator completed its order
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

			case consts.NewSlave:	// new elevator is trying to connect to the network and asking for role
				var ip string
				json.Unmarshal(typeJson.Data, &ip)
				clientConn := network.GetSendConn(ip)

				log.Println(consts.Cyan,"Recieved IP: ", ip, consts.Neutral)
				tmpDB := m.getDB()
				tmpList := tmpDB.getList()

				if tmpList.Len() == 2 {	// only master and new elevator are in the list => new elevator must be backup
					role := consts.Backup
					notification := consts.NotificationData{
						Code: consts.FindRole,
						Data: common.GetRawJSON(role),
					}
					m.sendToSlave(clientConn, notification)

				} else { // backup has been elected already
					role := consts.Slave
					notification := consts.NotificationData{
						Code: consts.FindRole,
						Data: common.GetRawJSON(role),
					}
					m.sendToSlave(clientConn, notification)
				}

			}
		}

	}
}

/**
 * Backup all info about connected elevators and hall orders to backup elevator.
 */
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

/**
 * Backup elevator became master => new connections are needed.
 */
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
 * Get first slave elevator from list. (for changing its role to backup)
 */
func (m *Master) getNewBackup() *net.UDPConn {
	slavesList := m.getDB().getList()
	if slavesList.Len() == 1 {
		return nil
	} else {
		for el := slavesList.Front(); el != nil; el = el.Next() {
			data := el.Value.(consts.DBItem)
			if data.Data.Role == consts.Slave {
				return data.ClientConn
			}
		}
	}
	return nil
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

		conn := master.getNewBackup()

		notification := consts.NotificationData{
			Code: consts.FindRole,
			Data: common.GetRawJSON(consts.Backup),
		}
		master.sendToSlave(conn, notification)

	} else {					// brand new master
		log.Println(consts.White, "Starting new master...", consts.Neutral)
	}

	go master.listenIncomingMsg(listenConn)
	go master.masterHallOrderHandler()
	go master.synchronizeDataWithBackup()
}






