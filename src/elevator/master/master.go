package master

import (
	"net"
	"sync"
	"consts"
	"time"
	"log"
	"encoding/json"
	"network"
	"elevator/common"
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

	data := common.GetNotification(notification)
	conn.Write(data)
}

func (m *Master) broadcastToSlaves(notification consts.NotificationData) {
	conn := network.GetSlaveTestSendConn() // TODO user actual broadcast function

	data := common.GetNotification(notification)
	//log.Println(consts.White, "broadcast", notification)

	conn.Write(data)
}

func (m *Master) masterOrderHandler() {
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
				item := elData.(dbItem)
				db.update(dbItem{
					ip: item.ip,
					ignore: 10,
					data: consts.PeriodicData{
						Floor:     item.data.Floor,
						Direction: item.data.Direction,
						CabArray:  item.data.CabArray,
						Free:      false,
					},
				})

				//ip := elData.(string) //TODO return non string type - net.Conn maybe?
				m.GetQueue().Pop()
				log.Println(consts.White, "parsed order", order, consts.Neutral)

				notification := consts.NotificationData{
					Code: consts.MasterHallOrder,
					Data: common.GetRawJSON(order),
				}

				// TODO get conn from "item.ip"
				m.sendToSlave(notification)
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
				data := consts.PeriodicData{} // for parsing of incoming message
				json.Unmarshal(typeJson.Data, &data)

				//queue := consts.Queue{}
				//err2 := json.Unmarshal(data.CabArray, &queue)
				//if err2 != nil {
				//	log.Println(consts.White, "array unmarshal master failed")
				//	log.Fatal(err2)
				//}

				//log.Printf("queue: %+v", queue)
				//log.Println(consts.White, "<- periodic", queue, consts.Neutral)
				//slaveData := SlaveData{
				//	Floor:      data.Floor,
				//	Direction:  data.Direction,
				//	OrderArray: data.CabArray,
				//	Free:      data.Free,
				//}

				//data.OrderArray = queue
				ip := getClientIPAddr()
				//m.GetDB().storeData(ip, slaveData)
				m.GetDB().storeData(ip, data)

				//m.GetDB().dump()
			case consts.SlaveHallOrder:
				order := consts.ButtonEvent{}
				json.Unmarshal(typeJson.Data, &order)
				log.Println(consts.White, "<- hall order", consts.Neutral)
				m.GetQueue().Push(order)

				// TODO broadcast all slaves to turn on light bulbs
				notification := consts.NotificationData{
					Code: consts.MasterHallLight,
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
func StartMaster(masterConn *net.UDPConn, listenConn *net.UDPConn) {
	common.ElevatorState.SetMasterConn(masterConn) // master conn has to be available for all roles

	slavesDB := SlavesDB{}
	master := Master{
		consts.NewQueue(),
		sync.Mutex{},
		&slavesDB,
	}

	go master.listenIncomingMsg(listenConn)
	go master.masterOrderHandler()
}






