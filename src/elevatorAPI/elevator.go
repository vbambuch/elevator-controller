package elevatorAPI

import (
	"consts"
	"network"
	"helper"
	"time"
	"sync"
)

/**
 *	Bunch of instruction creators.
 */
func MotorUp(stateChan chan<- []byte) ([]byte) {
	stateChan <- []byte{consts.MotorDirection, consts.MotorUp}
	return helper.GetInstruction(consts.MotorDirection, consts.MotorUp, consts.EmptyByte, consts.EmptyByte)
}

func MotorDown(stateChan chan<- []byte) ([]byte)  {
	stateChan <- []byte{consts.MotorDirection, consts.MotorDown}
	return helper.GetInstruction(consts.MotorDirection, consts.MotorDown, consts.EmptyByte, consts.EmptyByte)
}

func MotorStop(stateChan chan<- []byte) ([]byte)  {
	stateChan <- []byte{consts.MotorDirection, consts.MotorStop}
	return helper.GetInstruction(consts.MotorDirection, consts.MotorStop, consts.EmptyByte, consts.EmptyByte)
}

func FloorIndicator(floor int) ([]byte) {
	return helper.GetInstruction(consts.FloorIndicator, byte(floor), consts.EmptyByte, consts.EmptyByte)
}

func FloorSensor() ([]byte) {
	return helper.GetInstruction(consts.FloorSensor, consts.EmptyByte, consts.EmptyByte, consts.EmptyByte)
}



func stateHandler(stateChan <-chan []byte, instChan chan<- []byte, stateInfoChan chan<- consts.Elevator) {
	elevator := consts.Elevator{}

	for {
		buf := <- stateChan
		changed := false

		switch buf[0] {
		case consts.OrderButtonPressed:
			val := int(buf[1]) == 1
			if elevator.Data.OrderButton != val {
				elevator.Data.OrderButton = val
				changed = true
			}
		case consts.FloorSensor:
			val := int(buf[2])
			elevator.Data.AtFloor = int(buf[1]) == 1
			if elevator.Data.Floor != val && elevator.Data.AtFloor{
				elevator.Data.Floor = val
				changed = true

				instChan <- FloorIndicator(elevator.Data.Floor)
			}
		case consts.StopButtonPressed:
			val := int(buf[1]) == 1
			if elevator.Data.StopButton != val {
				elevator.Data.StopButton = val
				changed = true
			}
		case consts.ObstructionSwitch:
			val := int(buf[1]) == 1
			if elevator.Data.Obstruction != val {
				elevator.Data.Obstruction = val
				changed = true
			}
		case consts.MotorDirection:
			val := int(buf[1])
			if elevator.Data.Status != val {
				elevator.Data.Status = val
				changed = true
			}
		}


		if changed {
			elevator.Test = elevator.Data
			stateInfoChan <- elevator
		}
	}
}

func floorChecker(instrChan chan<- []byte) {
	for {
		instrChan <- FloorSensor()
		time.Sleep(500 * time.Millisecond)
	}
}

func Init() (chan []byte, chan []byte, chan consts.Elevator) {
	// channels
	stateChan := make(chan []byte)
	instrChan := make(chan []byte)
	stateInfoChan := make(chan consts.Elevator)

	socket := network.GetSocket(consts.Address, consts.Port)

	go floorChecker(instrChan)
	go stateHandler(stateChan, instrChan, stateInfoChan)
	go network.MessageReceiver(socket, stateChan)
	go network.MessageSender(socket, instrChan)

	return stateChan, instrChan, stateInfoChan
}
