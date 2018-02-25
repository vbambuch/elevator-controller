package tests

import (
	"consts"
	"helper"
	//"fmt"
)

func ZigZag(stateInfo <-chan consts.Elevator, instrChannel chan<- []byte) {
	for {
		info := <- stateInfo

		//fmt.Println("Got:", info)

		if info.AtFloor {
			switch info.Floor {
			case consts.MinFloor:
				instrChannel <- helper.GetInstruction(
					consts.MotorDirection, consts.MotorUp, consts.EmptyByte, consts.EmptyByte)
			case consts.MaxFloor:
				instrChannel <- helper.GetInstruction(
					consts.MotorDirection, consts.MotorDown, consts.EmptyByte, consts.EmptyByte)
			}
		}
	}
}
