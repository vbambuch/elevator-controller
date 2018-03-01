package elevator

import (
	"consts"
	"fmt"
	"time"
)


func stateHandler(floorChan <- chan int, obstructChan, stopChan <-chan bool, buttonsChan <- chan consts.ButtonEvent, stateChan chan Elevator)  {

	for {
		select {
		case button := <-buttonsChan:
			fmt.Printf("%+v\n", button)
			if button != ElevatorState.GetOrderButton() {
				ElevatorState.SetOrderButton(button)
				go SendElevatorToFloor(button.Floor, stateChan)
			}

		case floor := <-floorChan:
			fmt.Printf("floor: %+v\n", floor)
			if floor != ElevatorState.GetFloor() {
				if floor == consts.MinFloor || floor == consts.MaxFloor {
					ElevatorState.SetDirection(consts.MotorSTOP)
				}
				ElevatorState.SetFloorIndicator(floor)
			}

		case obstruct := <-obstructChan:
			fmt.Printf("%+v\n", obstruct)
			if obstruct != ElevatorState.GetObstruction() {
				ElevatorState.SetObstruction(obstruct)
			}

		case stop := <-stopChan:
			fmt.Printf("%+v\n", stop)
			if stop != ElevatorState.GetStopButton() {
				ElevatorState.SetStopButton(stop)

				for f := 0; f < consts.NumFloors; f++ {
					for b := consts.ButtonUP; b < consts.ButtonCAB; b++ {
						WriteButtonLamp(b, f, false)
					}
				}
			}
		}
		stateChan <- ElevatorState
	}
}

func SendElevatorToFloor(floor int, changeChan <-chan Elevator) {
	direction := consts.MotorUP

	if ElevatorState.floor > floor {
		direction = consts.MotorDOWN
	} else if ElevatorState.floor == floor {
		direction = consts.MotorSTOP
	}

	ElevatorState.SetDoorLight(false)
	ElevatorState.SetDirection(direction)

	for {
		info := <-changeChan

		if info.floor == floor {
			ElevatorState.SetDirection(consts.MotorSTOP)
			ElevatorState.SetDoorLight(true)
			ElevatorState.ClearOrderButton()
			return
		}
	}
}

func Init() (chan Elevator) {

	InitIO()

	buttonsChan := make(chan consts.ButtonEvent)
	floorChan := make(chan int)
	obstructChan := make(chan bool)
	stopChan := make(chan bool)
	stateChan := make(chan Elevator, 4)

	go PollButtons(buttonsChan)
	go PollFloorSensor(floorChan)
	go PollObstructionSwitch(obstructChan)
	go PollStopButton(stopChan)

	go stateHandler(floorChan, obstructChan, stopChan, buttonsChan, stateChan)

	// wait for initialization of elevator
	for ElevatorState.GetFloor() == -1 {
		ElevatorState.SetDirection(consts.MotorUP)
		time.Sleep(pollRate)
	}
	ElevatorState.SetDirection(consts.MotorSTOP)
	return stateChan
}
