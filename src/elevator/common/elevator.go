package common

import (
	"consts"
	"sync"
	"net"
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
	consts.ButtonEvent{consts.DefaultValue, consts.DefaultValue},
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
	hallOrder     	consts.ButtonEvent
	stopButton    	bool
	obstruction   	bool
	doorLight     	bool
	//hallQueue     *consts.Queue
	orderArray     []consts.ButtonEvent
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
		e.masterConn.Write(data)
		return true
	}
	return false
}


/**
 * Order array manipulation methods.
 */
func (e *Elevator) NewOrder(order consts.ButtonEvent) bool {
	e.mux.Lock()
	defer e.mux.Unlock()
	for _, v := range e.orderArray {
		if v.Floor == order.Floor && v.Button == order.Button { return false }
	}
	return true
}
func (e *Elevator) GetOrderArray() ([]consts.ButtonEvent) {
	e.mux.Lock()
	defer e.mux.Unlock()
	return e.orderArray
}

func (e *Elevator) OrderArrayNotEmpty() (bool) {
	return len(e.GetOrderArray()) != 0
}

// insert order to sorted list
func (e *Elevator) InsertToOrderArray(order consts.ButtonEvent) {
	e.mux.Lock()
	defer e.mux.Unlock()
	e.orderArray = append(e.orderArray, order)
}

// get first element regarding to direction of elevator
func (e *Elevator) GetOrder() (consts.ButtonEvent) {
	e.mux.Lock()
	defer e.mux.Unlock()

	// find out where elevator is going
	movingUP := e.direction == consts.MotorUP || e.prevDirection == consts.MotorUP
	movingDOWN := e.direction == consts.MotorDOWN || e.prevDirection == consts.MotorDOWN

	orderCount := len(e.orderArray)
	if orderCount == 1 {
		return e.orderArray[0]
	} else if movingUP {
		sort.Sort(helper.ASCFloors(e.orderArray))
	} else if movingDOWN {
		sort.Sort(helper.DESCFloors(e.orderArray))
	}

	for _, v := range e.orderArray {
		if (movingUP && v.Floor > e.floor) || (movingDOWN && v.Floor < e.floor) {
			// order is in the same direction as elevator
			return v
		} else if v.Floor == e.floor && !e.middleFloor {
			// order is from the same floor as elevator
			return v
		}
	}
	// all orders are in opposite direction => return last order (first in opposite direction)
	return e.orderArray[orderCount - 1]
}

func (e *Elevator) DeleteFirstElement() (interface{}) {
	var toRemove interface{}
	log.Println(consts.Yellow, "Prev order array:", ElevatorState.GetOrderArray(), consts.Neutral)

	e.mux.Lock()
	if e.OrderArrayNotEmpty() {
		toRemove = e.orderArray[0]
		e.orderArray = e.orderArray[1:]
	}
	e.mux.Unlock()

	log.Println(consts.Yellow, "Curr order array:", ElevatorState.GetOrderArray(), consts.Neutral)
	return toRemove
}

func (e *Elevator) DeleteOrder(order consts.ButtonEvent) {
	e.mux.Lock()
	for i, v := range e.orderArray {
		if v.Floor == order.Floor && v.Button == order.Button { // delete order
			e.orderArray = append(e.orderArray[:i], e.orderArray[i+1:]...)
		}
	}
	e.mux.Unlock()

	//log.Println(consts.Yellow, "Curr order array:", ElevatorState.GetOrderArray(), consts.Neutral)
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

func (e *Elevator) SetHallOrder(button consts.ButtonEvent) {
	e.mux.Lock()
	e.hallOrder = button
	e.mux.Unlock()
}

func (e *Elevator) ClearOrderButton(order consts.ButtonEvent) {
	WriteButtonLamp(order.Button, order.Floor, false)
}

func (e *Elevator) SetStopButton(stop bool, switchLamp bool) {
	e.mux.Lock()
	defer e.mux.Unlock()
	e.stopButton = stop
	if switchLamp {
		WriteStopLamp(stop)
	}
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
	if e.masterConn != nil {
		e.masterConn.Close()
	}
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
func (e *Elevator) IsMoving() bool {
	return e.GetDirection() != consts.MotorSTOP
}
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
func (e *Elevator) GetHallOrder() consts.ButtonEvent {
	e.mux.Lock()
	defer e.mux.Unlock()
	return e.hallOrder
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
//func (i *Elevator) InsertToOrderArray(qt consts.QueueType) *helper.Queue {
//	i.mux.Lock()
//	defer i.mux.Unlock()
//	queue := i.hallQueue
//	if qt == consts.OrderArray {
//		queue = i.orderArray
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
