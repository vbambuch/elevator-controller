package common

import (
	"consts"
	"time"
	"log"
)

func floorHandler(floorChan <-chan int) {
    for {
		floor := <-floorChan
		//log.Printf("floor: %+v\n", floor)
		if floor != ElevatorState.GetFloor() {
			if floor == consts.MinFloor || floor == consts.MaxFloor &&
				ElevatorState.GetDirection() != consts.MotorSTOP {
				ElevatorState.SetDirection(consts.MotorSTOP)
				ElevatorState.SetFloorIndicator(floor)
			} else if floor == consts.MiddleFloor {
				ElevatorState.SetMiddleFloor(true)
			} else {
				ElevatorState.SetMiddleFloor(false)
				ElevatorState.SetFloorIndicator(floor)
			}
		}
	}
}

func Init() (chan consts.ButtonEvent, chan bool, chan bool) {

	InitIO()

	buttonsChan := make(chan consts.ButtonEvent)
	floorChan := make(chan int)
	obstructChan := make(chan bool)
	stopChan := make(chan bool)

	go PollFloorSensor(floorChan)
	go PollObstructionSwitch(obstructChan)
	go PollStopButton(stopChan)
	go PollButtons(buttonsChan)

	go floorHandler(floorChan)

	// wait for initialization of elevator
	setup := true
	time.Sleep(2 * consts.PollRate) // wait for message exchange
	for ElevatorState.GetFloor() == consts.DefaultValue {
		if setup {
			ElevatorState.SetDirection(consts.MotorUP)
			log.Println(consts.Green, "Elevator is moving to floor...", consts.Neutral)
			setup = false
		}
		time.Sleep(consts.PollRate)
	}
	ElevatorState.SetDirection(consts.MotorSTOP)
	return buttonsChan, obstructChan, stopChan
}
