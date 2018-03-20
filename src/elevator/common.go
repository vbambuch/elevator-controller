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
	//helper.NewQueue(),
	sync.Mutex{},
	consts.Slave,
	nil,
}

type Elevator struct {
	direction     consts.MotorDirection
	prevDirection consts.MotorDirection
	floor         int
	prevFloor     int
	orderButton   consts.ButtonEvent
	stopButton    bool
	obstruction   bool
	doorLight     bool
	//hallQueue     *consts.Queue
	//cabQueue      *helper.Queue
	mux           sync.Mutex
	role 		  consts.Role
	masterConn	  *net.UDPConn
}

// Notifications
type Notification []byte
type notificationCode int

const (
	PeriodicNotify notificationCode = 0
	OrderNotify                     = 1
)

type NotificationData struct {
	Code        notificationCode
	Floor       int
	Direction   consts.MotorDirection
	OrderButton consts.ButtonEvent
	//CabQueue		*helper.Queue
}


/**
 * Common functions
 */
func (e *Elevator) sendToMaster(data Notification) {
	e.mux.Lock()
	defer e.mux.Unlock()
	e.masterConn.Write(data)
}

func (e *Elevator) periodicNotifications(finish <-chan bool) {
	for {
		fmt.Println("--> periodic notification")
		data := e.getNotification(PeriodicNotify)
		e.sendToMaster(data)
		time.Sleep(2 * time.Second)
	}
	<- finish
}

func (e *Elevator) orderNotifications(orderChan <-chan consts.ButtonEvent)  {
	for {
		order := <- orderChan
		fmt.Println("--> order", order)
		data := e.getNotification(OrderNotify)
		time.Sleep(2 * time.Second)
		e.sendToMaster(data)
	}
}

func (e *Elevator) listenFromMaster()  {
	conn := network.GetSlaveTestListenConn()
	var order consts.ButtonEvent //TODO change to more general
	buffer := make([]byte, 8192)

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

			fmt.Println("received from Master:", order)
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
	e.orderButton = button
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
func (e *Elevator) GetOrderButton() consts.ButtonEvent {
	e.mux.Lock()
	defer e.mux.Unlock()
	return e.orderButton
}
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
		e.orderButton,
	}

	//notification := NotificationData{e.floor, e.direction, e.cabQueue}
	data, err := json.Marshal(notification)
	helper.HandleError(err, "JSON error")

	return data
}
