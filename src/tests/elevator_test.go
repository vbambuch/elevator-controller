package tests

import (
	C "consts"
	"testing"
	//"bytes"
	//"helper"
	"elevator"
	"helper"
	"time"
)

/**
 * Basic elevator movement test
 */
func zigZag(t *testing.T, infoChan <-chan elevator.Elevator, stopChan chan<- bool) {

	go helper.Timeout(15000, stopChan)
	stuck := time.NewTimer(2 * time.Second)

	for {
		select {
		case info := <-infoChan:
			//fmt.Println("Got:", info)
			stuck.Reset(2 * time.Second)
			switch info.GetFloor() {
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
		case <-stuck.C:
			t.Error("Elevator is stuck.")
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
func makeOrder(floor int, bt C.ButtonType, stateInfoChan chan elevator.Elevator, readyChan chan bool)  {
	order := C.ButtonEvent{Floor: floor, Button: bt}
	elevator.ElevatorState.SetHallButton(order)
	go elevator.SendElevatorToFloor(order, stateInfoChan, readyChan)
	<- readyChan
	time.Sleep(2000 * time.Millisecond)
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

	makeOrder(2, C.ButtonDOWN, stateInfoChan, readyChan)
	makeOrder(3, C.ButtonDOWN, stateInfoChan, readyChan)
	makeOrder(0, C.ButtonUP, stateInfoChan, readyChan)
	makeOrder(3, C.ButtonDOWN, stateInfoChan, readyChan)

	return
}
