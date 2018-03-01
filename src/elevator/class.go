package elevator

import (
	"consts"
	//"fmt"
)

var ElevatorState = Elevator{}

type Elevator struct {
	Direction   	consts.MotorDirection
	PrevDirection	consts.MotorDirection
	Floor       	int
	PrevFloor       int
	OrderButton 	consts.ButtonEvent
	StopButton  	bool
	Obstruction 	bool
	DoorLight 		bool
}


func (i *Elevator) SetOrderButton(button consts.ButtonEvent) {
	i.OrderButton = button
	WriteButtonLamp(button.Button, button.Floor, true)
}

func (i *Elevator) SetFloorIndicator(floor int)  {
	if i.PrevFloor == -1 {
		i.PrevFloor = floor
	} else {
		i.PrevFloor = i.Floor
	}
	i.Floor = floor
	WriteFloorIndicator(floor)
}

func (i *Elevator) SetObstruction(obstruction bool)  {
	i.Obstruction = obstruction

	if obstruction {
		i.SetDirection(consts.MotorSTOP)
	} else {
		i.SetDirection(i.PrevDirection)
	}
}

func (i *Elevator) SetDirection(direction consts.MotorDirection)  {
	//fmt.Println("motor", direction)
	i.PrevDirection = i.Direction
	i.Direction = direction
	WriteMotorDirection(direction)
}

func (i *Elevator) SetStopButton(stop bool)  {
	i.StopButton = stop
}

func (i *Elevator) SetDoorLight(light bool)  {
	i.DoorLight = light
	WriteDoorOpenLamp(light)
}
