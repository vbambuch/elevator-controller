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
)


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

func (m *Master) sendToSlave(conn *net.UDPConn, notification consts.NotificationData) {
	m.mux.Lock()
	defer m.mux.Unlock()
	data := common.GetNotification(notification)
	if conn != nil {
		//log.Println(consts.White, "Send to:", conn.RemoteAddr(), consts.Neutral)
		conn.Write(data)
	}
}

func (m *Master) broadcastToSlaves(data consts.NotificationData) {
	//log.Println(consts.White, "broadcast", n)
	list := m.GetDB().getList()
	for el := list.Front(); el != nil; el = el.Next() {
		conn := el.Value.(consts.DBItem).ClientConn
		m.sendToSlave(conn, data)
	}
}

func (m *Master) masterHallOrderHandler() {
	db := m.GetDB()

	for {
		queue := m.GetQueue().Peek()
		if queue != nil { // empty queue test
			order := queue.(consts.ButtonEvent)
			//log.Println(consts.White, "peek of db", order)

			elData := db.findElevator(order)
			if elData != nil {
				//log.Println(consts.Red, "elData:", elData, consts.Neutral)

				// force this elevator to busy (don't wait for periodic update)
				// ignore next 5 updates from specific slave
				item := elData.(consts.DBItem)
				db.update(consts.DBItem{
					ClientConn: item.ClientConn,
					Ignore: 10,
					Data: consts.PeriodicData{
						ListenIP:       item.Data.ListenIP,
						Floor:          item.Data.Floor,
						Direction:      item.Data.Direction,
						OrderArray:     item.Data.OrderArray,
						Free:           false,
						HallProcessing: true,
					},
				})

				m.GetQueue().Pop()
				ip := item.Data.ListenIP
				log.Println(consts.White, ip, ": parsed order", order, consts.Neutral)
				//m.GetQueue().Dump()


				orderData := consts.NotificationData{
					Code: consts.MasterHallOrder,
					Data: common.GetRawJSON(order),
				}

				m.sendToSlave(item.ClientConn, orderData)
			} else {
				//log.Println(consts.White, "no free elevator")
			}
		}
		time.Sleep(consts.PollRate)
	}
}

func (m *Master) listenIncomingMsg(conn *net.UDPConn) {
	var typeJson consts.NotificationData
	buffer := make([]byte, 8192)
	hallQueue := m.GetQueue()

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

				m.GetDB().storeData(data)

				//m.GetDB().dump()
			case consts.SlaveHallOrder:
				order := consts.ButtonEvent{}
				json.Unmarshal(typeJson.Data, &order)
				log.Println(consts.White, "<- hall order", consts.Neutral)
				//log.Println(consts.White, "queue length:", hallQueue.Len(), consts.Neutral)

				if hallQueue.NewOrder(order) {
					hallQueue.Push(order)

					//m.GetQueue().Dump()

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
 * handle orders and store to queue
 * broadcast light indicators
 * sync DB with Backup
 * ping all slaves/backup
 * do same things as Slave
 */
func StartMaster() {
	listenConn := network.GetListenConn(consts.BListenAddress)

	slavesDB := SlavesDB{}
	master := Master{
		consts.NewQueue(),
		sync.Mutex{},
		&slavesDB,
	}

	go master.listenIncomingMsg(listenConn)
	go master.masterHallOrderHandler()
}






