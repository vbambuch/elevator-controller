package main

import (
	"fmt"
	"elevator"
	"network"
	"consts"
)

func roleDecision()  {
	role := network.FindOutRole()
	elevator.ElevatorState.SetRole(role)
}


func roleChangeHandler(orderChan <-chan consts.ButtonEvent, newRoleChan <-chan bool)  {
	started := false
	finish := make(chan bool)

	for {
		<- newRoleChan
		if started {
			finish <- true
		}
		started = true
		masterConn := network.GetMasterConn()
		listenConn := network.GetListenConn()
		switch elevator.ElevatorState.GetRole() {
		case consts.Master:
			fmt.Println("It's master")
			go elevator.StartMaster(orderChan, finish, masterConn, listenConn)
		case consts.Backup:
			fmt.Println("It's backup")
			go elevator.StartBackup(orderChan, finish, masterConn)
		case consts.Slave:
			fmt.Println("It's slave")
			go elevator.StartSlave(orderChan, finish, masterConn)
		}
	}
}

func errorHandler(errorChan <-chan consts.ElevatorError, newRoleChan chan<- bool) {
    for {
    	select {
		case err := <- errorChan:
			switch err.Code {
			case consts.MasterFailed:
				fmt.Println("Master failed")
				roleDecision()
				newRoleChan <- true
			case consts.BackupFailed:
				fmt.Println("Backup failed")
				roleDecision()
				newRoleChan <- true
			case consts.SlaveFailed:
				fmt.Println("Slave failed")
			}
		}
	}
}

func main() {
	errorChan := make(chan consts.ElevatorError)
	newRoleChan := make(chan bool)
	//stateChan := make(chan elevator.Elevator)
	orderChan := make(chan consts.ButtonEvent)

	// master-backup-slave decision
	roleDecision()

	// initiate specific
	//elevator.Init(stateChan, orderChan)
	elevator.Init(orderChan)

	// start error detection
	go network.ErrorDetection(errorChan)
	go errorHandler(errorChan, newRoleChan)
	go roleChangeHandler(orderChan, newRoleChan)

	newRoleChan <- true
	fmt.Println("App started")
	blocker := make(chan bool, 1)
	<- blocker
}
