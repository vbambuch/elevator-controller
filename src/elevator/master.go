package elevator

import (
	"net"
	"sync"
	"fmt"
	"consts"
	"time"
	"log"
	"encoding/json"
	"network"
	"helper"
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

//func (m *Master) sendToSlave(ip string, order consts.ButtonEvent)  { //TODO uncomment when ready
func (m *Master) sendToSlave(order consts.ButtonEvent)  {
	conn := network.GetSlaveTestSendConn() // TODO get conn from somewhere

	data, err := json.Marshal(order) // TODO move to common functions
	helper.HandleError(err, "JSON error")

	conn.Write(data)
}

func (m *Master) masterOrderHandler()  {
	for {
		dbData := m.GetQueue().Pop()
		if dbData != nil {
			order := dbData.(consts.ButtonEvent)
			fmt.Printf("[master] poped from db: %+v\n", order)

			elData := m.GetDB().findFreeElevator(order.Floor)
			if elData != nil {
				//ip := elData.(string) //TODO return not string - net.Conn maybe?
				fmt.Printf("[master] parsed order: %+v\n", order)
				m.sendToSlave(order)
			} else {
				fmt.Println("[master] no free elevator")
			}
		}
		
		
		
		time.Sleep(pollRate)
	}
}

func (m *Master) listenIncomingMsg(conn *net.UDPConn)  {
	var data NotificationData
	buffer := make([]byte, 8192)

	for {
		n, err := conn.Read(buffer[0:])
		if err != nil {
			fmt.Println("reading master failed")
			log.Fatal(err)
		}
		/*fmt.Println(string(buffer))*/
		//fmt.Println(buffer)
		if len(buffer) > 0 {
			//fmt.Println(string(buffer))
			err2 := json.Unmarshal(buffer[0:n], &data)
			if err2 != nil {
				fmt.Println("unmarshal master failed")
				log.Fatal(err2)
			}

			//fmt.Println("received", data)


			switch data.Code {
			case PeriodicMsg:
				//fmt.Println("-------------------------------")
				fmt.Println("[master] <- periodic")
				//fmt.Println("floor:", data.Floor)
				//fmt.Println("direction:", data.Direction)
				//fmt.Println("-------------------------------")
				ip := getClientIPAddr()
				m.GetDB().storeData(ip, data)

			case HallOrderMsg:
				fmt.Println("[master] <- hall order")
				//fmt.Println("-------------------------------")
				//fmt.Print("received order:", data.HallOrder)
				//fmt.Println("-------------------------------")
				order := data.HallOrder
				m.GetQueue().Push(order)
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
	data 		NotificationData
}

type SlavesDB struct {
	array 	[]dbItem
	mux 	sync.Mutex
}

func (i *SlavesDB) exists(ip string) (bool) {
	for _, v := range i.array {
		if v.ip == ip {
			return true
		}
	}
	return false
}

func (i *SlavesDB) update(item dbItem) {
	for _, v := range i.array {
		if v.ip == item.ip {
			queue := v.data.CabQueue // keep previous queue
			v.data = item.data
			v.data.CabQueue = queue
		}
	}
}

func (i *SlavesDB) storeData(ip string, data NotificationData)  {
	i.mux.Lock()
	defer i.mux.Unlock()

	item := dbItem{ip, data}
	if i.exists(ip) {
		i.update(item)
	} else {
		i.array = append(i.array, item)
	}

	//fmt.Printf("db: %+v\n", i)
}

func (i *SlavesDB) findElevatorsOnFloor(floor int) interface{} {
	for _, v := range i.array {
		elData := v.data
		if elData.Floor == floor {
			return v
		}
	}
	return nil
}

func (i *SlavesDB) findFreeElevator(floor int) interface{} {
	for _, v := range i.array {
		return v.ip
	}
	return nil
}
