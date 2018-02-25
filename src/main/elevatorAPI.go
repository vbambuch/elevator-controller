package main

import (
	"fmt"
	"time"
	"consts"
	"network"
	"tests"
	"helper"
)

func stateHandler(stateChange <-chan []byte, instChannel chan<- []byte, stateInfo chan<- consts.Elevator) {
	elevator := consts.Elevator{}

	for {
		buf := <- stateChange
		changed := false

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

				instChannel <- helper.GetInstruction(
					consts.FloorIndicator, byte(elevator.Floor), consts.EmptyByte, consts.EmptyByte)
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
		}

		if changed {
			stateInfo <- elevator
		}
	}
}

func floorChecker(instrChannel chan<- []byte) {
    for {
		instrChannel <- helper.GetInstruction(
			consts.FloorSensor, consts.EmptyByte, consts.EmptyByte, consts.EmptyByte)
		time.Sleep(500 * time.Millisecond)
	}
}



func main() {
	// channels
	stateChange := make(chan []byte)
	instrChannel := make(chan []byte)
	stateInfo := make(chan consts.Elevator)


	fmt.Println("connection created")
	socket := network.GetSocket(consts.Address, consts.Port)

	go floorChecker(instrChannel)
	go stateHandler(stateChange, instrChannel, stateInfo)
	go network.ReceiveMessage(socket, stateChange)
	go network.SendMessage(socket, instrChannel)

	go tests.ZigZag(stateInfo, instrChannel)

	instrChannel <- helper.GetInstruction(
		consts.MotorDirection, consts.MotorUp, consts.EmptyByte, consts.EmptyByte)

	blocker := make(chan bool, 1)
	<- blocker
}
