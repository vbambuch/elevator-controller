package elevator

import (
	"consts"
	//"fmt"
	"sync"
	"helper"
	//"fmt"
	"encoding/json"
	"fmt"
	"time"
	"net"
	//"log"
	"network"
	"log"
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
}

// Notifications
type Notification []byte
type notificationCode int

const (
	PeriodicMsg  notificationCode = 0
	HallOrderMsg                  = 1
)

type NotificationData struct {
	Code      notificationCode
	Floor     int
	Direction consts.MotorDirection
	HallOrder consts.ButtonEvent
	CabQueue  *consts.Queue
}


/**
 * Common functions
 */
func (e *Elevator) sendToMaster(data Notification) bool {
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
		data := e.getNotification(PeriodicMsg)
		if e.sendToMaster(data) {
			fmt.Println("[slave] -> periodic")
		}
		time.Sleep(20 * time.Second)
	}
}

func (e *Elevator) HallOrderNotifications(sendHallChan <-chan consts.ButtonEvent)  {
	for {
		order := <-sendHallChan
		data := e.getNotification(HallOrderMsg)
		if e.sendToMaster(data) {
	 		fmt.Println("[slave] -> hall order:", order)
		}
	}
}

func (e *Elevator) ListenFromMaster(receivedHallChan chan<- consts.ButtonEvent) {
	conn := network.GetSlaveTestListenConn()
	var order consts.ButtonEvent //TODO change to more general
	buffer := make([]byte, 8192)
	//receivedOrder := make(chan consts.ButtonEvent)

	for {
		n, err := conn.Read(buffer[0:])
		if err != nil {
			fmt.Println("reading slave failed")
			log.Fatal(err)
		}
		/*fmt.Println(string(buffer))*/
		//fmt.Println(buffer)
		if len(buffer) > 0 {
			//fmt.Println(string(buffer))
			err2 := json.Unmarshal(buffer[0:n], &order)
			if err2 != nil {
				fmt.Println("unmarshal slave failed")
				log.Fatal(err2)
			}

			fmt.Println("[slave] <- hall order:", order)
			receivedHallChan <- order
		}
	}
}

func (e *Elevator) OrderHandler(cabButtonChan <-chan consts.ButtonEvent, hallButtonChan <-chan consts.ButtonEvent)  {
	var timeout = time.NewTimer(0)
	ready := false
	onFloorChan := make(chan bool)
	//floorChan := make(chan int)

	for {
		select {
		case <- onFloorChan:
			timeout.Reset(3 * time.Second)
		case <- timeout.C:
			fmt.Println("Elevator ready")
			ElevatorState.SetDoorLight(false)
			ready = true
		case cabOrder := <- cabButtonChan:
			if ready {
				fmt.Printf("Ready for cab %d\n", cabOrder.Floor)
				go SendElevatorToFloor(cabOrder, onFloorChan)
				ready = false
			} else {
				fmt.Printf("Pushed to cab queue %+v\n", cabOrder)
				ElevatorState.GetQueue().Push(cabOrder)
			}

		case hallOrder := <- hallButtonChan:
			if ready {
				fmt.Printf("Ready for hall %d\n", hallOrder.Floor)
				go SendElevatorToFloor(hallOrder, onFloorChan)
				ready = false
			} else {
				fmt.Printf("Pushed to hall queue %+v\n", hallOrder)
				//ElevatorState.GetQueue(consts.HallQueue).Push(order)
			}

		//case state := <-orderChan:
		//	fmt.Println("New state")
		//
		//
		//	if order.Button != consts.DefaultValue {
		//		fmt.Println("not default")
		//
		//		floor := state.GetFloor()
		//		floorChan <- floor
		//		fmt.Println("after")
		//
		//	}


		default:
			//if ElevatorState.GetQueue(consts.HallQueue).Len() != 0 &&
			//	ElevatorState.GetQueue(consts.CabQueue).Len() == 0 &&
			//	ready {
			//	// pop order from hall queue
			//	queueOrder := ElevatorState.GetQueue(consts.HallQueue).Pop().(consts.ButtonEvent)
			//	fmt.Printf("Pop from hall queue %+v\n", queueOrder)
			//	go SendElevatorToFloor(queueOrder, floorChan, onFloorChan)
			//	ready = false

			if ElevatorState.GetQueue().Len() != 0 && ready {
				// pop order from cab queue
				queueOrder := ElevatorState.GetQueue().Pop().(consts.ButtonEvent)
				fmt.Printf("Pop from cab queue %+v\n", queueOrder)
				go SendElevatorToFloor(queueOrder, onFloorChan)
				ready = false
			}
		}

	}
}





/**
 * Bunch of setters.
 */
func (e *Elevator) SetDirection(direction consts.MotorDirection) {
	//fmt.Println("motor", direction)
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

func (e *Elevator) SetOrderButton(button consts.ButtonEvent) {
	e.mux.Lock()
	//e.hallOrder = button
	e.mux.Unlock()
	WriteButtonLamp(button.Button, button.Floor, true)
}

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
//	if qt == consts.CabQueue {
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

func (e *Elevator) getNotification(code notificationCode) (Notification) {
	e.mux.Lock()
	defer e.mux.Unlock()
	notification := NotificationData{
		code,
		e.floor,
		e.direction,
		e.hallOrder,
		nil,
	}

	//notification := NotificationData{e.floor, e.direction, e.cabQueue}
	data, err := json.Marshal(notification)
	helper.HandleError(err, "JSON error")

	return data
}
