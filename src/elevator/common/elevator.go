package common

import (
	"consts"
	"sync"
	"encoding/json"
	"time"
	"net"
	"network"
	"log"
	//"helper"
	//"container/list"
	"sort"
	"helper"
)

// ElevatorState constructor
var ElevatorState = Elevator {
	consts.MotorSTOP,
	consts.MotorSTOP,
	consts.DefaultValue,
	consts.DefaultValue,
	false,
	//consts.ButtonEvent{consts.DefaultValue, consts.DefaultValue},
	false,
	false,
	false,
	//helper.NewQueue(),
	[]consts.ButtonEvent{},
	sync.Mutex{},
	consts.Slave,
	nil,
	true,
	false,
}

type Elevator struct {
	direction     	consts.MotorDirection
	prevDirection 	consts.MotorDirection
	floor         	int
	prevFloor     	int
	middleFloor		bool
	//hallOrder     	consts.ButtonEvent
	stopButton    	bool
	obstruction   	bool
	doorLight     	bool
	//hallQueue     *consts.Queue
	cabArray       []consts.ButtonEvent
	mux            sync.Mutex
	role           consts.Role
	masterConn     *net.UDPConn
	free           bool
	hallProcessing bool
}


/**
 * Common functions
 */
func (e *Elevator) sendToMaster(data consts.Notification) bool {
	e.mux.Lock()
	defer e.mux.Unlock()

	if e.masterConn != nil {
		e.masterConn.Write(data) // TODO change masterConn when Master will change
		return true
	}
	return false
}

func (e *Elevator) PeriodicNotifications() {
	for {
		data := consts.PeriodicData{
			Floor:          e.floor,
			Direction:      e.direction,
			CabArray:       e.cabArray,
			Free:           e.free,
			HallProcessing: e.hallProcessing,
		}
		notification := consts.NotificationData{
			Code: consts.SlavePeriodicMsg,
			Data: GetRawJSON(data),
		}

		msg := GetNotification(notification)
		if e.sendToMaster(msg) {
			//log.Println(consts.Blue, "-> periodic", *e.cabArray, consts.Neutral)
		}
		//time.Sleep(1 * time.Second)
		time.Sleep(consts.PollRate)
	}
}

func (e *Elevator) HallOrderNotifications(sendHallChan <-chan consts.ButtonEvent)  {
	for {
		order := <-sendHallChan
		notification := consts.NotificationData{
			Code: consts.SlaveHallOrder,
			Data: GetRawJSON(order),
		}

		msg := GetNotification(notification)
		if e.sendToMaster(msg) {
	 		log.Println(consts.Blue, "-> hall order:", order, consts.Neutral)
		}
	}
}

func (e *Elevator) ListenIncomingMsg(receivedHallChan chan<- consts.ButtonEvent) {
	conn := network.GetSlaveTestListenConn()
	var typeJson consts.NotificationData
	buffer := make([]byte, 8192)
	//receivedOrder := make(chan consts.ButtonEvent)

	for {
		n, err := conn.Read(buffer[0:])
		if err != nil {
			log.Println(consts.Blue, "reading slave failed", consts.Neutral)
			log.Fatal(err)
		}
		//log.Println(consts.Blue, buffer, consts.Neutral)
		if len(buffer) > 0 {
			//log.Println(consts.Blue, string(buffer), consts.Neutral)
			err2 := json.Unmarshal(buffer[0:n], &typeJson)
			if err2 != nil {
				log.Println(consts.Blue, "unmarshal slave failed", consts.Neutral)
				log.Fatal(err2)
			} else {

				//log.Println(consts.Blue, "<- received typeJson", typeJson, consts.Neutral)

				switch typeJson.Code {
				case consts.MasterHallOrder:
					order := consts.ButtonEvent{}
					json.Unmarshal(typeJson.Data, &order)
					log.Println(consts.Blue, "<- hall order:", order, consts.Neutral)
					receivedHallChan <- order
				case consts.MasterHallLight:
					order := consts.ButtonEvent{}
					json.Unmarshal(typeJson.Data, &order)
					log.Println(consts.Blue, "<- hall light:", order, consts.Neutral)
					WriteButtonLamp(order.Button, order.Floor, true)
				case consts.MasterBroadcastIP:
					var ip string
					json.Unmarshal(typeJson.Data, &ip)
					log.Println(consts.Blue, "<- master ip:", ip, consts.Neutral)

				}
			}
		}
	}
}

func (e *Elevator) OrderHandler(cabButtonChan <-chan consts.ButtonEvent, hallButtonChan <-chan consts.ButtonEvent)  {
	var timeout = time.NewTimer(0)
	free := false
	cabInterrupted := false
	onFloorChan := make(chan bool)
	interruptCab := make(chan bool)

	for {
		select {
		case <- onFloorChan:
			timeout.Reset(3 * time.Second)
		case <- timeout.C:
			ElevatorState.SetDoorLight(false)
			ElevatorState.SetFree(true)
			free = true

			// elevator is ready for another hall call
			if ElevatorState.GetHallProcessing() {
				ElevatorState.SetHallProcessing(false)
			}

		case cabOrder := <- cabButtonChan:
			if !ElevatorState.OrderExists(cabOrder) {
				if ElevatorState.GetDirection() != consts.MotorSTOP {
					interruptCab <- true
					cabInterrupted = true
				}
				log.Println(consts.Blue, "Pushed to cab queue", cabOrder, consts.Neutral)
				ElevatorState.InsertToCabArray(cabOrder)
				log.Println(consts.Yellow, "Curr cab array:", ElevatorState.GetCabArray(), consts.Neutral)
			}

		case hallOrder := <- hallButtonChan:
			ElevatorState.SetHallProcessing(true)
			if free {
				log.Println(consts.Blue, "Free for hall", hallOrder.Floor, consts.Neutral)
				go SendElevatorToFloor(hallOrder, onFloorChan, interruptCab)
				free = false
			} else {
				interruptCab <- true
				log.Println(consts.Blue, "Interrupt and hall", hallOrder.Floor, consts.Neutral)
				go SendElevatorToFloor(hallOrder, onFloorChan, interruptCab)
				free = false
			}

		default:
			if len(ElevatorState.GetCabArray()) != 0 && (free || cabInterrupted) {
				// get first cab order
				queueOrder := ElevatorState.GetCabOrder()
				log.Println(consts.Blue, "Read from cab queue", queueOrder, consts.Neutral)
				go SendElevatorToFloor(queueOrder, onFloorChan, interruptCab)
				free = false
				cabInterrupted = false
			}
		}
	}
}





/**
 * List manipulation methods.
 */
func (e *Elevator) OrderExists(order consts.ButtonEvent) bool {
	e.mux.Lock()
	defer e.mux.Unlock()
	for _, v := range e.cabArray {
		if v.Floor == order.Floor { return true }
	}
	return false
}
func (e *Elevator) GetCabArray() ([]consts.ButtonEvent) {
	e.mux.Lock()
	defer e.mux.Unlock()
	return e.cabArray
}

// insert order to sorted list
func (e *Elevator) InsertToCabArray(order consts.ButtonEvent) {
	e.mux.Lock()
	defer e.mux.Unlock()
	e.cabArray = append(e.cabArray, order)
}

// get first element regarding to direction of elevator
func (e *Elevator) GetCabOrder() (consts.ButtonEvent) {
	e.mux.Lock()
	defer e.mux.Unlock()

	// find out where elevator is going
	movingUP := e.direction == consts.MotorUP || e.prevDirection == consts.MotorUP
	movingDOWN := e.direction == consts.MotorDOWN || e.prevDirection == consts.MotorDOWN

	orderCount := len(e.cabArray)

	if orderCount == 1 {
		return e.cabArray[0]
	} else if movingUP {
		sort.Sort(helper.ASCFloors(e.cabArray))
	} else if movingDOWN {
		sort.Sort(helper.DESCFloors(e.cabArray))
	}

	for _, v := range e.cabArray {
		if (movingUP && v.Floor > e.floor) || (movingDOWN && v.Floor < e.floor) {
			// order is in the same direction as elevator
			return v
		} else if v.Floor == e.floor && !e.middleFloor {
			// order is from the same floor as elevator
			return v
		}
	}
	// all orders are in opposite direction => return last order (first in opposite direction)
	return e.cabArray[orderCount - 1]
}

func (e *Elevator) DeleteFirstElement() (interface{}) {
	var toRemove interface{}
	log.Println(consts.Yellow, "Prev cab array:", ElevatorState.GetCabArray(), consts.Neutral)

	e.mux.Lock()
	if len(e.cabArray) != 0 {
		toRemove = e.cabArray[0]
		e.cabArray = e.cabArray[1:]
	}
	e.mux.Unlock()

	log.Println(consts.Yellow, "Curr cab array:", ElevatorState.GetCabArray(), consts.Neutral)
	return toRemove
}

func (e *Elevator) DeleteOrder(order consts.ButtonEvent) {
	//log.Println(consts.Yellow, "Prev cab array:", ElevatorState.GetCabArray(), consts.Neutral)

	e.mux.Lock()
	for i, v := range e.cabArray {
		if v.Floor == order.Floor { // delete order
			e.cabArray = append(e.cabArray[:i], e.cabArray[i+1:]...)
		}
	}
	e.mux.Unlock()

	//log.Println(consts.Yellow, "Curr cab array:", ElevatorState.GetCabArray(), consts.Neutral)
}


/**
 * Bunch of setters.
 */
func (e *Elevator) SetDirection(direction consts.MotorDirection) {
	//log.Println(consts.Blue, "motor", direction, consts.Neutral)
	e.mux.Lock()
	e.prevDirection = e.direction
	e.direction = direction
	e.mux.Unlock()
	WriteMotorDirection(direction)
}

func (e *Elevator) SetFloorIndicator(floor int) {
	e.mux.Lock()
	if e.prevFloor == consts.DefaultValue {
		e.prevFloor = floor
	} else {
		e.prevFloor = e.floor
	}
	e.floor = floor
	e.mux.Unlock()
	WriteFloorIndicator(floor)
}

func (e *Elevator) SetMiddleFloor(a bool) {
	e.mux.Lock()
	defer e.mux.Unlock()
	e.middleFloor = a
}

//func (e *Elevator) SetHallButton(button consts.ButtonEvent) {
//	e.mux.Lock()
//	e.hallOrder = button
//	e.mux.Unlock()
//}

func (e *Elevator) ClearOrderButton(order consts.ButtonEvent) {
	WriteButtonLamp(order.Button, order.Floor, false)
}

func (e *Elevator) SetStopButton(stop bool) {
	e.mux.Lock()
	defer e.mux.Unlock()
	e.stopButton = stop
}

func (e *Elevator) SetObstruction(obstruction bool) {
	e.mux.Lock()
	e.obstruction = obstruction

	if obstruction {
		e.SetDirection(consts.MotorSTOP)
	} else {
		e.SetDirection(e.prevDirection)
	}
	e.mux.Unlock()
}

func (e *Elevator) SetDoorLight(light bool) {
	e.mux.Lock()
	e.doorLight = light
	e.mux.Unlock()
	WriteDoorOpenLamp(light)
}

func (e *Elevator) SetRole(role consts.Role) {
	e.mux.Lock()
	defer e.mux.Unlock()
	e.role = role
}

func (e *Elevator) SetMasterConn(conn *net.UDPConn) {
	e.mux.Lock()
	defer e.mux.Unlock()
	e.masterConn = conn
}

func (e *Elevator) SetFree(free bool) {
	e.mux.Lock()
	defer e.mux.Unlock()
	e.free = free
}

func (e *Elevator) SetHallProcessing(processing bool) {
	e.mux.Lock()
	defer e.mux.Unlock()
	e.hallProcessing = processing
}



/**
 * Bunch of getters
 */
func (e *Elevator) GetDirection() consts.MotorDirection {
	e.mux.Lock()
	defer e.mux.Unlock()
	return e.direction
}
func (e *Elevator) GetPrevDirection() consts.MotorDirection {
	e.mux.Lock()
	defer e.mux.Unlock()
	return e.prevDirection
}
func (e *Elevator) GetFloor() int {
	e.mux.Lock()
	defer e.mux.Unlock()
	return e.floor
}
func (e *Elevator) GetPrevFloor() int {
	e.mux.Lock()
	defer e.mux.Unlock()
	return e.prevFloor
}
func (e *Elevator) IsMiddleFloor() bool {
	e.mux.Lock()
	defer e.mux.Unlock()
	return e.middleFloor
}
//func (e *Elevator) GetOrderButton() consts.ButtonEvent {
//	e.mux.Lock()
//	defer e.mux.Unlock()
//	return e.hallOrder
//}
func (e *Elevator) GetStopButton() bool {
	e.mux.Lock()
	defer e.mux.Unlock()
	return e.stopButton
}
func (e *Elevator) GetObstruction() bool {
	e.mux.Lock()
	defer e.mux.Unlock()
	return e.obstruction
}
func (e *Elevator) GetDoorLight() bool {
	e.mux.Lock()
	defer e.mux.Unlock()
	return e.doorLight
}
//func (i *Elevator) InsertToCabArray(qt consts.QueueType) *helper.Queue {
//	i.mux.Lock()
//	defer i.mux.Unlock()
//	queue := i.hallQueue
//	if qt == consts.CabArray {
//		queue = i.cabArray
//	}
//	return queue
//}


func (e *Elevator) GetRole() (consts.Role) {
	e.mux.Lock()
	defer e.mux.Unlock()
	return e.role
}

func (e *Elevator) GetMasterConn() (*net.UDPConn) {
	e.mux.Lock()
	defer e.mux.Unlock()
	return e.masterConn
}

func (e *Elevator) GetFree() (bool) {
	e.mux.Lock()
	defer e.mux.Unlock()
	return e.free
}

func (e *Elevator) GetHallProcessing() (bool) {
	e.mux.Lock()
	defer e.mux.Unlock()
	return e.hallProcessing
}
