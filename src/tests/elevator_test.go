package tests

import (
	C "consts"
	"testing"
	//"bytes"
	//"helper"
	"elevator"
	"fmt"
	"helper"
	"time"
)

/**
 * Basic elevator movement test
 */
func zigZag(t *testing.T, infoChan <-chan elevator.Elevator, stopChan chan<- bool) {
	go helper.Timeout(15000, stopChan)

	for {
		info := <-infoChan
		//fmt.Println("Got:", info)

		switch elevator.ReadFloor() {
		//switch info.floor {
		case C.MinFloor:
			if info.GetDirection() != C.MotorSTOP && info.GetPrevFloor() != 0 {
				t.Errorf("Motor should go 0, not %d", info.GetDirection())
				stopChan <- true
			}
			if info.GetFloor() != C.MinFloor {
				t.Errorf("Indicator should be 0, not %d", info.GetFloor())
				stopChan <- true
			}
			elevator.ElevatorState.SetDirection(C.MotorUP)
		case 1:
			if info.GetFloor() != 1 {
				t.Errorf("Indicator should be 1, not %d", info.GetFloor())
				stopChan <- true
			}
		case 2:
			if info.GetFloor() != 2 {
				t.Errorf("Indicator should be 2, not %d", info.GetFloor())
				stopChan <- true
			}
		case C.MaxFloor:
			if info.GetDirection() != C.MotorSTOP && info.GetPrevFloor() != 0 {
				t.Errorf("Motor should go 0, not %d", info.GetDirection())
				stopChan <- true
			}
			if info.GetFloor() != C.MaxFloor {
				t.Errorf("Indicator should be 3, not %d", info.GetFloor())
				stopChan <- true
			}
			elevator.ElevatorState.SetDirection(C.MotorDOWN)
		}
	}

}

func TestZigZag(t *testing.T) {
	// TODO check if elevator is actually moving
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
/*
	- get order from PollButton
	- send elevator to correct direction
	- stop it when elevator reaches its destination
	- open door -> turn on the light
	- wait for some time OR cab call
	- turn off door light
	- send elevator to destination
	- open door
	- wait for some time OR hall call OR cab call
 */

	readyChan := make(chan bool)
	stateInfoChan := elevator.Init()

	elevator.ElevatorState.SetOrderButton(C.ButtonEvent{2, C.ButtonType(1)})
	go elevator.SendElevatorToFloor(2, stateInfoChan, readyChan)
	<- readyChan
	time.Sleep(2000 * time.Millisecond)
	elevator.ElevatorState.SetOrderButton(C.ButtonEvent{3, C.ButtonType(1)})
	go elevator.SendElevatorToFloor(3, stateInfoChan, readyChan)
	<- readyChan
	time.Sleep(2000 * time.Millisecond)
	elevator.ElevatorState.SetOrderButton(C.ButtonEvent{0, C.ButtonType(0)})
	go elevator.SendElevatorToFloor(0, stateInfoChan, readyChan)
	<- readyChan
	time.Sleep(2000 * time.Millisecond)
	elevator.ElevatorState.SetOrderButton(C.ButtonEvent{3, C.ButtonType(1)})
	go elevator.SendElevatorToFloor(3, stateInfoChan, readyChan)
	<- readyChan
	time.Sleep(2000 * time.Millisecond)

	return
}
