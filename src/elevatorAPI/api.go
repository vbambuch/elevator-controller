package elevatorAPI

import (
	"consts"
	"helper"
)

/**
 *	Bunch of instruction creators.
 */
func WriteMotorUp(stateChan chan<- []byte) ([]byte) {
	stateChan <- []byte{consts.MotorDirection, consts.MotorUp}
	return helper.GetInstruction(consts.MotorDirection, consts.MotorUp, consts.EmptyByte, consts.EmptyByte)
}

func WriteMotorDown(stateChan chan<- []byte) ([]byte)  {
	stateChan <- []byte{consts.MotorDirection, consts.MotorDown}
	return helper.GetInstruction(consts.MotorDirection, consts.MotorDown, consts.EmptyByte, consts.EmptyByte)
}

func WriteMotorStop(stateChan chan<- []byte) ([]byte)  {
	stateChan <- []byte{consts.MotorDirection, consts.MotorStop}
	return helper.GetInstruction(consts.MotorDirection, consts.MotorStop, consts.EmptyByte, consts.EmptyByte)
}

func WriteFloorIndicator(floor int) ([]byte) {
	return helper.GetInstruction(consts.FloorIndicator, byte(floor), consts.EmptyByte, consts.EmptyByte)
}


func ReadFloorSensor() ([]byte) {
	return helper.GetInstruction(consts.FloorSensor, consts.EmptyByte, consts.EmptyByte, consts.EmptyByte)
}

func ReadOrderButton(button, floor byte) ([]byte) {
	return helper.GetInstruction(consts.OrderButtonPressed, button, floor, consts.EmptyByte)
}
