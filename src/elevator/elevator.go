package elevator

import (
	"consts"
	//"helper"
	"fmt"
	"reflect"
)



func shareElevatorStatus(stateChan <-chan Elevator, changeChan chan<- Elevator) {
    prevState := Elevator{}

    for {
    	state := <-stateChan
		if !reflect.DeepEqual(prevState, state) {
			fmt.Printf("changed %+v\n", state)
			prevState = state
			changeChan <- state
		} else {
			//fmt.Printf("same %+v\n", state)
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
			//fmt.Printf("%+v\n", floor)
			if (floor == consts.MinFloor || floor == consts.MaxFloor) && ElevatorState.PrevFloor != 0  {
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
				for b := consts.ButtonType(0); b < 3; b++ {
					WriteButtonLamp(b, f, false)
				}
			}
		}
		stateChan <- ElevatorState
	}
}

func Init() (chan Elevator) {

	InitIO()

	buttonsChan := make(chan consts.ButtonEvent)
	floorChan := make(chan int)
	obstructChan := make(chan bool)
	stopChan := make(chan bool)
	stateChan := make(chan Elevator)
	changeChan := make(chan Elevator)

	go PollButtons(buttonsChan)
	go PollFloorSensor(floorChan)
	go PollObstructionSwitch(obstructChan)
	go PollStopButton(stopChan)

	go stateHandler(floorChan, obstructChan, stopChan, buttonsChan, stateChan)
	go shareElevatorStatus(stateChan, changeChan)

	return changeChan
}
