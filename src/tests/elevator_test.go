package tests

import (
	C "consts"
	"testing"
	"bytes"
	"helper"
	"elevatorAPI"
	//"time"
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
func zigZag(stateInfoChan <-chan C.Elevator, instrChan chan<- []byte, stateChan chan<- []byte) {
	instrChan <- elevatorAPI.MotorUp(stateChan)

	for {
		info := <- stateInfoChan

		fmt.Println("Got data:", info.Data)

		if info.Data.AtFloor {
			switch info.Data.Floor {
			case C.MinFloor:
				instrChan <- elevatorAPI.MotorUp(stateChan)
			case C.MaxFloor:
				instrChan <- elevatorAPI.MotorDown(stateChan)
			}
		}
	}
}

func floorTesting(stateInfoChan <-chan C.Elevator, stopChan chan<- bool, t *testing.T)  {
	//time.Sleep(10000 * time.Millisecond)
	floor := -1

	for {
		info := <- stateInfoChan

		fmt.Println("Got test:", info.Test)

		if info.Test.AtFloor {
			if floor != -1 {
				switch floor {
				case 0:
					if info.Test.Status != C.MotorUp {
						t.Errorf("Motor should go UP, not %d", info.Test.Status)
						stopChan <- true
					}
				case 3:
					if info.Test.Status != C.MotorDown {
						t.Errorf("Motor should go DOWN, not %d", info.Test.Status)
						stopChan <- true
					}
				}
			} else {
				floor = info.Test.Floor
			}
		}
	}

	stopChan <- true
}

func TestZigZag(t *testing.T) {
	stateChan, instrChan, stateInfoChan := elevatorAPI.Init()
	stopChan := make(chan bool)

	go zigZag(stateInfoChan, instrChan, stateChan)
	go floorTesting(stateInfoChan, stopChan, t)

	//fmt.Println("Test started")
	<- stopChan
}


