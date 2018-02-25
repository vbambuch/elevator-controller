package main

import (
	"fmt"
	"net"
	"time"
	"consts"
	"network"
	"helper"
)

type State struct {
	floor int
	atFloor bool
	status int
	orderButton bool
	stopButton bool
	obstruction bool
}


func getInstruction(b0, b1, b2, b3 byte) ([]byte){
	return []byte{b0, b1, b2, b3}

}


func stateHandler(stateChange <-chan []byte, instChannel chan<- []byte, stateInfo chan<- State) {
	state := State{}

	for {
		buf := <- stateChange

		switch buf[0] {
		case consts.OrderButtonPressed:
			state.orderButton = int(buf[1]) == 1
		case consts.FloorSensor:
			state.atFloor = int(buf[1]) == 1
			state.floor = int(buf[2])
			if state.atFloor {
				instChannel <- getInstruction(consts.FloorIndicator, byte(state.floor), consts.EmptyByte, consts.EmptyByte)
			}
		case consts.StopButtonPressed:
			state.stopButton = int(buf[1]) == 1
		case consts.ObstructionSwitch:
			state.obstruction = int(buf[1]) == 1
		}

		stateInfo <- state
	}
}

func floorChecker(instrChannel chan<- []byte) {
    for {
		instrChannel <- getInstruction(consts.FloorSensor, consts.EmptyByte, consts.EmptyByte, consts.EmptyByte)
		time.Sleep(500 * time.Millisecond)
	}
}

func zigZag(stateInfo <-chan State, instrChannel chan<- []byte) {
	for {
		info := <- stateInfo

		if info.atFloor {
			switch info.floor {
			case consts.MinFloor:
				instrChannel <- getInstruction(consts.MotorDirection, consts.MotorUp, consts.EmptyByte, consts.EmptyByte)
			case consts.MaxFloor:
				instrChannel <- getInstruction(consts.MotorDirection, consts.MotorDown, consts.EmptyByte, consts.EmptyByte)
			}
		}
	}
}


func main() {
	// channels
	stateChange := make(chan []byte)
	instrChannel := make(chan []byte)
	stateInfo := make(chan State)

	// Set up send socket
	addr, err := net.ResolveTCPAddr("tcp", ":15657")
	helper.HandleError(err, "Resolve tcp")

	socket, err := net.DialTCP("tcp", nil, addr)
	helper.HandleError(err, "Dial tcp")

	fmt.Println("connection created")


	go floorChecker(instrChannel)
	go stateHandler(stateChange, instrChannel, stateInfo)
	go network.ReceiveMessage(socket, stateChange)
	go network.SendMessage(socket, instrChannel)

	go zigZag(stateInfo, instrChannel)

	instrChannel <- getInstruction(consts.MotorDirection, consts.MotorUp, consts.EmptyByte, consts.EmptyByte)

	blocker := make(chan bool, 1)
	<- blocker
}
