package common

import (
	"consts"
	"time"
	"log"
	"network"
)

func floorHandler(floorChan <-chan int) {
    for {
		floor := <-floorChan
		//log.Printf("floor changed: %+v\n", floor)
		if floor == consts.ElevatorFailed {
			ElevatorState.SetMasterConn(nil)
			ReInitIO()
			addr := network.GetBroadcastAddress()+consts.MasterPort
			ElevatorState.SetMasterConn(network.GetSendConn(addr))
			ElevatorState.SetFree(true)
			//} else if floor != ElevatorState.GetFloor() {
		} else {
			if floor == consts.MinFloor || floor == consts.MaxFloor &&
				ElevatorState.IsMoving() {
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

func initializeElevator() {
	setup := true
	time.Sleep(2 * consts.PollRate) // wait for message exchange

	log.Println(consts.Cyan, "Floor init:", ElevatorState.GetFloor(), consts.Neutral)

	for ElevatorState.GetFloor() == consts.MiddleFloor || ElevatorState.GetFloor() == consts.DefaultValue {
		if setup {
			ElevatorState.SetDirection(consts.MotorUP)
			log.Println(consts.Green, "Elevator is moving to floor...", consts.Neutral)
			setup = false
		}
		time.Sleep(consts.PollRate)
	}
	ElevatorState.SetDirection(consts.MotorSTOP)
}

func defaultElevatorState()  {
	for f := 0; f <= consts.MaxFloor; f++ {
		for b := consts.ButtonType(0); b < 3; b++ {
			WriteButtonLamp(b, f, false)
		}
	}
	WriteDoorOpenLamp(false)
	WriteStopLamp(false)
}

func Init() (chan consts.ButtonEvent, chan bool, chan bool) {

	InitIO()

	buttonsChan := make(chan consts.ButtonEvent)
	floorChan := make(chan int)
	obstructChan := make(chan bool)
	stopChan := make(chan bool)

	go pollFloorSensor(floorChan)
	go pollObstructionSwitch(obstructChan)
	go pollStopButton(stopChan)
	go pollButtons(buttonsChan)

	go floorHandler(floorChan)

	// clear all call, door and stop buttons
	defaultElevatorState()

	// wait for initialization of elevator
	initializeElevator()
	return buttonsChan, obstructChan, stopChan
}
