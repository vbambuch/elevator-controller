package main

import (
	"network"
	"consts"
	"log"
	"elevator/master"
	"elevator/common"
	"elevator/slave"
)

func roleDecision()  {
	role := network.FindOutRole()
	//role := network.FindOutNewMaster()
	common.ElevatorState.SetRole(role)
	//common.ElevatorState.SetMasterConn()
}


func startCommonProcedures(receivedCabChan <-chan consts.ButtonEvent, sendHallChan <-chan consts.ButtonEvent)  {
	receivedHallChan := make(chan consts.ButtonEvent)

	go common.ElevatorState.PeriodicNotifications()
	go common.ElevatorState.HallOrderNotifications(sendHallChan)
	go common.ElevatorState.ListenIncomingMsg(receivedHallChan)
	go common.ElevatorState.OrderHandler(receivedCabChan, receivedHallChan)
}

func roleChangeHandler(orderChan <-chan consts.ButtonEvent, newRoleChan <-chan bool)  {
	//started := false
	//finish := make(chan bool)

	for {
		<- newRoleChan
		//if started {
		//	finish <- true
		//}
		//started = true
		masterConn := network.GetMasterConn()
		listenConn := network.GetListenConn()
		switch common.ElevatorState.GetRole() {
		case consts.Master:
			log.Println(consts.Blue, "It's master", consts.Neutral)
			go master.StartMaster(masterConn, listenConn)
		case consts.Backup:
			log.Println(consts.Blue, "It's backup", consts.Neutral)
			go slave.StartBackup(orderChan, masterConn)
		case consts.Slave:
			log.Println(consts.Blue, "It's slave", consts.Neutral)
			go slave.StartSlave(orderChan, masterConn)
		}
	}
}

func errorHandler(errorChan <-chan consts.ElevatorError, newRoleChan chan<- bool) {
    for {
    	select {
		case err := <- errorChan:
			switch err.Code {
			case consts.MasterFailed:
				log.Println("Master failed")
				roleDecision()
				newRoleChan <- true
			case consts.BackupFailed:
				log.Println("Backup failed")
				roleDecision()
				newRoleChan <- true
			case consts.SlaveFailed:
				log.Println("Slave failed")
			}
		}
	}
}

//Channels for the network
var outgoingMsg = make(chan consts.Message, 10)
var incomingMsg = make(chan consts.Message, 10)

func main() {
	errorChan := make(chan consts.ElevatorError)
	newRoleChan := make(chan bool)
	//stateChan := make(chan common.Elevator)
	orderChan := make(chan consts.ButtonEvent)

	// master-backup-slave decision
	roleDecision()

	// initiate specific


	//common.Init(stateChan, orderChan)
	cabButtonChan, hallButtonChan := common.Init()
	network.Initialize(outgoingMsg, incomingMsg)

	// start error detection
	//go network.ErrorDetection(errorChan)
	go errorHandler(errorChan, newRoleChan)
	go startCommonProcedures(cabButtonChan, hallButtonChan)
	go roleChangeHandler(orderChan, newRoleChan)

	newRoleChan <- true
	log.Println(consts.Green, "App started", consts.Neutral)
	blocker := make(chan bool, 1)
	<- blocker
}
