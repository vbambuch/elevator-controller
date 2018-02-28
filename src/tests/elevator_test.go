package tests

import (
	C "consts"
	"testing"
	//"bytes"
	//"helper"
	"elevator"
	"fmt"
	"helper"
)

/**
 * Basic dumb test
 */
//type byteTable struct { a byte; b byte; c byte; d byte; r []byte }

//func TestGetInstr(t *testing.T)  {
//	table := []byteTable{
//		{C.MotorDirection, C.MotorUp, C.EmptyByte, C.EmptyByte, []byte{1, 100, 0, 0}},
//		{C.OrderButtonLight, C.CabButton, byte(3), C.TurnOn, []byte{2, 2, 3, 1}},
//	}
//
//	for _, value := range table {
//		instr := helper.GetInstruction(value.a, value.b, value.c, value.d)
//		if bytes.Compare(instr, value.r) != 0 {
//			t.Errorf("Incorrect instruction, got: %d, want: %d.", instr, value.r)
//		}
//	}
//}

/**
 * Basic elevator movement test
 */
func zigZag(t *testing.T, infoChan <-chan elevator.Elevator, stopChan chan<- bool) {
	go helper.Timeout(150000, stopChan)

	for {
		info := <-infoChan
		//fmt.Println("Got:", info)

		switch elevator.ReadFloor() {
		//switch info.Floor {
		case C.MinFloor:
			if info.Direction != C.MotorSTOP && info.PrevFloor != 0 {
				t.Errorf("Motor should go 0, not %d", info.Direction)
				stopChan <- true
			}
			if info.Floor != C.MinFloor {
				t.Errorf("Indicator should be 0, not %d", info.Floor)
				stopChan <- true
			}
			elevator.ElevatorState.SetDirection(C.MotorUP)
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
			if info.Direction != C.MotorSTOP && info.PrevFloor != 0 {
				t.Errorf("Motor should go 0, not %d", info.Direction)
				stopChan <- true
			}
			if info.Floor != C.MaxFloor {
				t.Errorf("Indicator should be 3, not %d", info.Floor)
				stopChan <- true
			}
			elevator.ElevatorState.SetDirection(C.MotorDOWN)
		}
	}

}

func TestZigZag(t *testing.T) {
	stateInfoChan := elevator.Init()

	stopChan := make(chan bool)

	elevator.ElevatorState.SetDirection(C.MotorUP)
	go zigZag(t, stateInfoChan, stopChan)

	<- stopChan
}


/**
 * Elevator calls testing
 */
func elevatorCalls(stateInfoChan <-chan elevator.Elevator)  {
	for {
		info := <- stateInfoChan
		fmt.Printf("%+v\n", info)
		//if info.AtFloor {
		//}
	}
}


func TestElevatorCalls(t *testing.T)  {
	stateInfoChan := elevator.Init()
	stopChan := make(chan bool)

	go elevatorCalls(stateInfoChan)

	<- stopChan
}
