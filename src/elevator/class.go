package elevator

import (
	"consts"
	//"fmt"
	"sync"
)

// ElevatorState constructor
var ElevatorState = Elevator{
	3,
	3,
	-1,
	-1,
	consts.ButtonEvent{-1, -1},
	false,
	false,
	false,
	sync.Mutex{},
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
	mux           sync.Mutex
}

/**
 * Bunch of setters.
 */
func (i *Elevator) SetDirection(direction consts.MotorDirection)  {
	//fmt.Println("motor", direction)
	i.mux.Lock()
	i.prevDirection = i.direction
	i.direction = direction
	i.mux.Unlock()
	WriteMotorDirection(direction)
}

func (i *Elevator) SetFloorIndicator(floor int)  {
	i.mux.Lock()
	if i.prevFloor == -1 {
		i.prevFloor = floor
	} else {
		i.prevFloor = i.floor
	}
	i.floor = floor
	i.mux.Unlock()
	WriteFloorIndicator(floor)
}

func (i *Elevator) SetOrderButton(button consts.ButtonEvent) {
	i.mux.Lock()
	i.orderButton = button
	i.mux.Unlock()
	WriteButtonLamp(button.Button, button.Floor, true)
}

func (i *Elevator) ClearOrderButton() {
	i.mux.Lock()
	WriteButtonLamp(i.orderButton.Button, i.orderButton.Floor, false)
	i.mux.Unlock()
}

func (i *Elevator) SetStopButton(stop bool)  {
	i.mux.Lock()
	i.stopButton = stop
	i.mux.Unlock()
}

func (i *Elevator) SetObstruction(obstruction bool)  {
	i.mux.Lock()
	i.obstruction = obstruction

	if obstruction {
		i.SetDirection(consts.MotorSTOP)
	} else {
		i.SetDirection(i.prevDirection)
	}
	i.mux.Unlock()
}

func (i *Elevator) SetDoorLight(light bool)  {
	i.mux.Lock()
	i.doorLight = light
	i.mux.Unlock()
	WriteDoorOpenLamp(light)
}


/**
 * Bunch of getters
 */
func (i *Elevator) GetDirection() (consts.MotorDirection) {
	i.mux.Lock()
	defer i.mux.Unlock()
	return i.direction
}
func (i *Elevator) GetPrevDirection() (consts.MotorDirection) {
	i.mux.Lock()
	defer i.mux.Unlock()
	return i.prevDirection
}
func (i *Elevator) GetFloor() (int) {
	i.mux.Lock()
	defer i.mux.Unlock()
	return i.floor
}
func (i *Elevator) GetPrevFloor() (int) {
	i.mux.Lock()
	defer i.mux.Unlock()
	return i.prevFloor
}
func (i *Elevator) GetOrderButton() (consts.ButtonEvent) {
	i.mux.Lock()
	defer i.mux.Unlock()
	return i.orderButton
}
func (i *Elevator) GetStopButton() (bool) {
	i.mux.Lock()
	defer i.mux.Unlock()
	return i.stopButton
}
func (i *Elevator) GetObstruction() (bool) {
	i.mux.Lock()
	defer i.mux.Unlock()
	return i.obstruction
}
func (i *Elevator) GetDoorLight() (bool) {
	i.mux.Lock()
	defer i.mux.Unlock()
	return i.doorLight
}
