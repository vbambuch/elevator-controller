package tests

import (
	C "consts"
	"testing"
	"bytes"
	"helper"
	"elevatorAPI"
	"fmt"
)

/**
 * Basic dumb test
 */
type byteTable struct { a byte; b byte; c byte; d byte; r []byte }

func TestGetInstr(t *testing.T)  {
	table := []byteTable{
		{C.MotorDirection, C.MotorUp, C.EmptyByte, C.EmptyByte, []byte{1, 100, 0, 0}},
		{C.OrderButtonLight, C.CabButton, byte(3), C.TurnOn, []byte{2, 2, 3, 1}},
	}

	for _, value := range table {
		instr := helper.GetInstruction(value.a, value.b, value.c, value.d)
		if bytes.Compare(instr, value.r) != 0 {
			t.Errorf("Incorrect instruction, got: %d, want: %d.", instr, value.r)
		}
	}
}

/**
 * Basic elevator movement test
 */
func zigZag(t *testing.T, stateInfoChan <-chan C.Elevator, instrChan chan<- []byte, stateChan chan<- []byte, stopChan chan<- bool) {
	instrChan <- elevatorAPI.WriteMotorUp(stateChan)

	go helper.Timeout(15000, stopChan)

	for {
		info := <- stateInfoChan
		//fmt.Println("Got:", info)

		if info.AtFloor {
			switch info.Floor {
			case C.MinFloor:
				if info.Status != C.MotorDown {
					t.Errorf("Motor should go 200, not %d", info.Status)
					stopChan <- true
				}
				if info.Floor != C.MinFloor {
					t.Errorf("Indicator should be 0, not %d", info.Floor)
					stopChan <- true
				}
				instrChan <- elevatorAPI.WriteMotorUp(stateChan)
			case 1:
				if info.Floor != 1 {
					t.Errorf("Indicator should be 1, not %d", info.Floor)
					stopChan <- true
				}
			case 2:
				if info.Floor != 2 {
					t.Errorf("Indicator should be 2, not %d", info.Floor)
					stopChan <- true
				}
			case C.MaxFloor:
				if info.Status != C.MotorUp {
					t.Errorf("Motor should go 100, not %d", info.Status)
					stopChan <- true
				}
				if info.Floor != C.MaxFloor {
					t.Errorf("Indicator should be 3, not %d", info.Floor)
					stopChan <- true
				}
				instrChan <- elevatorAPI.WriteMotorDown(stateChan)
			}
		}
	}

}

func TestZigZag(t *testing.T) {
	stateChan, instrChan, stateInfoChan := elevatorAPI.Init()
	stopChan := make(chan bool)

	go zigZag(t, stateInfoChan, instrChan, stateChan, stopChan)

	<- stopChan
}


/**
 * Elevator calls testing
 */
func elevatorCalls(stateInfoChan <-chan C.Elevator)  {
	for {
		info := <- stateInfoChan
		fmt.Println("Got:", info)
		//if info.AtFloor {
		//}
	}
}


func TestElevatorCalls(t *testing.T)  {
	_, _, stateInfoChan := elevatorAPI.Init()
	stopChan := make(chan bool)

	go elevatorCalls(stateInfoChan)

	<- stopChan
}
