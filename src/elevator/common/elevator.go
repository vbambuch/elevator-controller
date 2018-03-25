package common

import (
	"consts"
	"sync"
	"encoding/json"
	"time"
	"net"
	"network"
	"log"
	"helper"
)

// ElevatorState constructor
var ElevatorState = Elevator {
	consts.MotorSTOP,
	consts.MotorSTOP,
	consts.DefaultValue,
	consts.DefaultValue,
	consts.ButtonEvent{consts.DefaultValue, consts.DefaultValue},
	false,
	false,
	false,
	//helper.NewQueue(),
	consts.NewQueue(),
	sync.Mutex{},
	consts.Slave,
	nil,
	true,
}

type Elevator struct {
	direction     consts.MotorDirection
	prevDirection consts.MotorDirection
	floor         int
	prevFloor     int
	hallOrder     consts.ButtonEvent
	stopButton    bool
	obstruction   bool
	doorLight     bool
	//hallQueue     *consts.Queue
	cabQueue      *consts.Queue
	mux           sync.Mutex
	role 		  consts.Role
	masterConn	  *net.UDPConn
	ready		  bool
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
			Floor:     e.floor,
			Direction: e.direction,
			CabArray:  helper.QueueToArray(*e.cabQueue),
			Ready:     e.ready,
		}
		notification := consts.NotificationData{
			Code: consts.SlavePeriodicMsg,
			Data: GetRawJSON(data),
		}

		msg := GetNotification(notification)
		if e.sendToMaster(msg) {
			//log.Println(consts.Blue, "-> periodic", *e.cabQueue, consts.Neutral)
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
	ready := false
	popCabCall := true
	onFloorChan := make(chan bool)
	interruptCab := make(chan bool)

	for {
		select {
		case <- onFloorChan:
			if popCabCall {
				order := ElevatorState.GetQueue().Pop()
				log.Println(consts.Blue, "Clear cab order", order, consts.Neutral)
			} else {
				popCabCall = true
			}
			timeout.Reset(3 * time.Second)
		case <- timeout.C:
			//log.Println(consts.Blue, "Elevator ready", consts.Neutral)
			ElevatorState.SetDoorLight(false)
			ElevatorState.SetReady(true)
			ready = true
		case cabOrder := <- cabButtonChan:
			//if ready {
			//	log.Println(consts.Blue, "Ready for cab", cabOrder.Floor, consts.Neutral)
			//	ElevatorState.GetQueue().Push(cabOrder)
			//	go SendElevatorToFloor(cabOrder, onFloorChan)
			//	ready = false
			//} else {
				log.Println(consts.Blue, "Pushed to cab queue", cabOrder, consts.Neutral)
			// TODO ElevatorState.PushToQueue(cabOrder) + check if cab order exists
			// TODO sort cab calls by floor => not queue

				ElevatorState.GetQueue().Push(cabOrder)
			//}

		case hallOrder := <- hallButtonChan:
			if ready {
				log.Println(consts.Blue, "Ready for hall", hallOrder.Floor, consts.Neutral)
				go SendElevatorToFloor(hallOrder, onFloorChan, interruptCab)
				ready = false
			} else {

				log.Println(consts.Blue, "Interrupt and hall", hallOrder.Floor, consts.Neutral)
				go SendElevatorToFloor(hallOrder, onFloorChan, interruptCab)
				popCabCall = false
				ready = false
				//log.Println(consts.Red, "Slave received another hall queue", hallOrder, consts.Neutral)
				//log.Println(consts.Blue, "Pushed to hall queue", hallOrder, consts.Neutral)
				//ElevatorState.GetQueue(consts.HallQueue).Push(order)
			}

		//case state := <-orderChan:
		//	log.Println(consts.Blue, "New state", consts.Neutral)
		//
		//
		//	if order.Button != consts.DefaultValue {
		//		log.Println(consts.Blue, "not default", consts.Neutral)
		//
		//		floor := state.GetFloor()
		//		floorChan <- floor
		//		log.Println(consts.Blue, "after", consts.Neutral)
		//
		//	}


		default:
			//if ElevatorState.GetQueue(consts.HallQueue).Len() != 0 &&
			//	ElevatorState.GetQueue(consts.CabArray).Len() == 0 &&
			//	ready {

			//	// pop order from hall queue
			//	queueOrder := ElevatorState.GetQueue(consts.HallQueue).Pop().(consts.ButtonEvent)
			//	log.Println(consts.Blue, "Pop from hall queue", queueOrder, consts.Neutral)
			//	go SendElevatorToFloor(queueOrder, floorChan, onFloorChan)
			//	ready = false

			if ElevatorState.GetQueue().Len() != 0 && ready {
				// pop order from cab queue
				queueOrder := ElevatorState.GetQueue().Peek().(consts.ButtonEvent)
				log.Println(consts.Blue, "Read from cab queue", queueOrder, consts.Neutral)
				go SendElevatorToFloor(queueOrder, onFloorChan, interruptCab)
				ready = false
			}
		}

	}
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
	if e.prevFloor == -1 {
		e.prevFloor = floor
	} else {
		e.prevFloor = e.floor
	}
	e.floor = floor
	e.mux.Unlock()
	WriteFloorIndicator(floor)
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

func (e *Elevator) SetReady(ready bool) {
	e.mux.Lock()
	defer e.mux.Unlock()
	e.ready = ready
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
//func (i *Elevator) GetQueue(qt consts.QueueType) *helper.Queue {
//	i.mux.Lock()
//	defer i.mux.Unlock()
//	queue := i.hallQueue
//	if qt == consts.CabArray {
//		queue = i.cabQueue
//	}
//	return queue
//}

func (e *Elevator) GetQueue() *consts.Queue {
	e.mux.Lock()
	defer e.mux.Unlock()
	return e.cabQueue
}

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

func (e *Elevator) GetReady() (bool) {
	e.mux.Lock()
	defer e.mux.Unlock()
	return e.ready
}
