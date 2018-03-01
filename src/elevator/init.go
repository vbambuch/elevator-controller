package elevator

import (
	"consts"
	"fmt"
	"reflect"
	"time"
)



func shareElevatorStatus(stateChan <-chan Elevator, changeChan chan<- Elevator) {
    prevState := Elevator{}

    for {
    	state := <-stateChan
		if !reflect.DeepEqual(prevState, state) {
			fmt.Printf("changed %+v\n", state)
			prevState = state
			changeChan <- state
		}
	}
}

func stateHandler(floorChan <- chan int, obstructChan, stopChan <-chan bool, buttonsChan <- chan consts.ButtonEvent, stateChan chan<- Elevator)  {

	for {
		select {
		case button := <-buttonsChan:
			//fmt.Printf("%+v\n", button)
			ElevatorState.SetOrderButton(button)

		case floor := <-floorChan:
			//fmt.Printf("floor: %+v\n", floor)
			if floor == consts.MinFloor || floor == consts.MaxFloor {
				ElevatorState.SetDirection(consts.MotorSTOP)
			}
			ElevatorState.SetFloorIndicator(floor)

		case obstruct := <-obstructChan:
			//fmt.Printf("%+v\n", obstruct)
			ElevatorState.SetObstruction(obstruct)

		case stop := <-stopChan:
			//fmt.Printf("%+v\n", stop)
			ElevatorState.SetStopButton(stop)

			for f := 0; f < consts.NumFloors; f++ {
				for b := consts.ButtonUP; b < consts.ButtonCAB; b++ {
					WriteButtonLamp(b, f, false)
				}
			}
		}
		stateChan <- ElevatorState
	}
}

func SendElevatorToFloor(floor int, changeChan <-chan Elevator) {
	direction := consts.MotorUP

	if ElevatorState.Floor > floor {
		direction = consts.MotorDOWN
	} else if ElevatorState.Floor == floor {
		direction = consts.MotorSTOP
	}

	ElevatorState.SetDoorLight(false)
	ElevatorState.SetDirection(direction)

	for {
		info := <-changeChan

		if info.Floor == floor {
			ElevatorState.SetDirection(consts.MotorSTOP)
			ElevatorState.SetDoorLight(true)
			return
		}
	}
}

func Init() (chan Elevator) {

	InitIO()

	// setup default ElevatorState properties
	ElevatorState.Floor = -1
	ElevatorState.PrevFloor = -1
	ElevatorState.DoorLight = false

	buttonsChan := make(chan consts.ButtonEvent)
	floorChan := make(chan int)
	obstructChan := make(chan bool)
	stopChan := make(chan bool)
	stateChan := make(chan Elevator)
	changeChan := make(chan Elevator, 2)

	go PollButtons(buttonsChan)
	go PollFloorSensor(floorChan)
	go PollObstructionSwitch(obstructChan)
	go PollStopButton(stopChan)

	go stateHandler(floorChan, obstructChan, stopChan, buttonsChan, stateChan)
	go shareElevatorStatus(stateChan, changeChan)

	// wait for initialization of elevator
	for ElevatorState.Floor == -1 {
		ElevatorState.SetDirection(consts.MotorUP)
		time.Sleep(pollRate)
	}
	return changeChan
}
