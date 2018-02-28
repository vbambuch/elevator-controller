package elevatorAPI

import (
	"consts"
	"network"
	//"helper"
	"time"
)



func stateHandler(stateChan <-chan []byte, instChan chan<- []byte, stateInfoChan chan<- consts.Elevator) {
	elevator := consts.Elevator{}

	for {
		buf := <- stateChan
		changed := false //for checking previous value

		switch buf[0] {
		case consts.OrderButtonPressed:
			val := int(buf[1]) == 1
			if elevator.OrderButton != val {
				elevator.OrderButton = val
				changed = true
			}
		case consts.FloorSensor:
			val := int(buf[2])
			elevator.AtFloor = int(buf[1]) == 1
			if elevator.Floor != val && elevator.AtFloor{
				elevator.Floor = val
				changed = true

				instChan <- WriteFloorIndicator(elevator.Floor)
			}
		case consts.StopButtonPressed:
			val := int(buf[1]) == 1
			if elevator.StopButton != val {
				elevator.StopButton = val
				changed = true
			}
		case consts.ObstructionSwitch:
			val := int(buf[1]) == 1
			if elevator.Obstruction != val {
				elevator.Obstruction = val
				changed = true
			}
		case consts.MotorDirection:
			val := int(buf[1])
			if elevator.Status != val {
				elevator.Status = val
				//changed = true
			}
		}

		if changed {
			stateInfoChan <- elevator
		}
	}
}

func floorSensorChecker(instrChan chan<- []byte) {
	for {
		instrChan <- ReadFloorSensor()
		time.Sleep(500 * time.Millisecond)
	}
}

func orderButtonChecker(instrChan chan<- []byte) {
	for {
		instrChan <- ReadOrderButton(consts.CabButton, byte(2))
		//for floor := 0; floor <= 3; floor++ {
		//	for button := 0; button <= 2; button++ {
		//	}
		//}
		time.Sleep(500 * time.Millisecond)
	}
}

func Init() (chan []byte, chan []byte, chan consts.Elevator) {
	// channels
	stateChan := make(chan []byte)
	instrChan := make(chan []byte)
	stateInfoChan := make(chan consts.Elevator)

	socket := network.GetSocket(consts.Address, consts.Port)

	go network.MessageReceiver(socket, stateChan)
	go network.MessageSender(socket, instrChan)

	go floorSensorChecker(instrChan)
	go orderButtonChecker(instrChan)
	go stateHandler(stateChan, instrChan, stateInfoChan)

	return stateChan, instrChan, stateInfoChan
}
