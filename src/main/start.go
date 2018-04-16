package main

import (
	"network"
	"consts"
	"log"
	"elevator/master"
	"elevator/common"
	"elevator/slave"
	"flag"
)

func roleDecision()  {
	role := network.FindOutRole()
	//role := network.FindOutNewMaster()
	common.ElevatorState.SetRole(role)
	//common.ElevatorState.SetMasterConn()
}

/**
 * Run several goroutines that are the same for every role
 */
func startCommonProcedures(
	buttonsChan <-chan consts.ButtonEvent,
	obstructChan <-chan bool,
	stopChan <-chan bool)  {

	receivedHallChan := make(chan consts.ButtonEvent)

	masterConn := network.GetSendConn(network.GetBroadcastAddress()+consts.MasterPort)
	common.ElevatorState.SetMasterConn(masterConn)

	ipAddr := network.IncreasePortForAddress(masterConn.LocalAddr().String())
	conn := network.GetListenConn(ipAddr)

	go common.PeriodicNotifications(ipAddr)
	go common.ListenIncomingMsg(receivedHallChan, conn)
	go common.ButtonsHandler(buttonsChan, receivedHallChan, obstructChan, stopChan)
}

func resolveRoleChange(masterFailed chan<- consts.BackupSync, backupData consts.BackupSync)  {
	switch common.ElevatorState.GetRole() {
	case consts.Master:
		log.Println(consts.Blue, "It's master", consts.Neutral)
		go master.StartMaster(backupData)
	case consts.Backup:
		log.Println(consts.Blue, "It's backup", consts.Neutral)
		go slave.StartBackup(masterFailed)
	case consts.Slave:
		log.Println(consts.Blue, "It's slave", consts.Neutral)
	}
}

func roleChangeHandler(newRoleChan <-chan bool)  {
	//started := false
	//finish := make(chan bool)
	masterFailed := make(chan consts.BackupSync)
	backupData := consts.BackupSync{}

	for {
		select {
		case <- newRoleChan:
			resolveRoleChange(masterFailed, backupData)
		case backupData = <- masterFailed:
			log.Println(consts.Yellow, "backup data:", backupData, consts.Neutral)
			common.ElevatorState.SetRole(consts.Master)
			resolveRoleChange(masterFailed, backupData)
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

func main() {
	numFloors := flag.Int("numFloors", 4, "elevator id")
	id := flag.Int("id", 0, "elevator id")
	elPort := flag.String("elPort", "15657", "my elevator port")

	// TODO remove when network is done...
	//masterPort := flag.String("masterPort", "20002", "master localhost port")
	//myPort := flag.String("myPort", "20000", "my localhost port")
	//myRole := flag.Int("myRole", 1, "1: Master, 2: Backup, 3: Slave")
	flag.Parse()

	//log.Println(consts.Green, "Master port:", *masterPort, consts.Neutral)
	//log.Println(consts.Green, "My port:", *myPort, consts.Neutral)

	//log.Println(consts.Green, "My role:", *myRole, consts.Neutral)

	consts.ElevatorPort = *elPort
	//consts.MasterPort = *masterPort
	//consts.MyPort = *myPort

	// TODO ...remove when network is done


	log.Println(consts.Green, "Elevator ID:", *id, consts.Neutral)
	log.Println(consts.Green, "Number of floors:", *numFloors, consts.Neutral)
	log.Println(consts.Green, "Elevator port:", *elPort, consts.Neutral)

	consts.NumFloors = *numFloors
	consts.MaxFloor = *numFloors - 1

	errorChan := make(chan consts.ElevatorError)
	newRoleChan := make(chan bool)

	// master-backup-slave decision
	roleDecision()	// TODO uncomment
	//common.ElevatorState.SetRole(consts.Role(*myRole)) // TODO remove

	buttonsChan, obstructChan, stopChan := common.Init()
	//network.Initialize(outgoingMsg, incomingMsg)

	// start error detection
	//go network.ErrorDetection(errorChan)
	go errorHandler(errorChan, newRoleChan)
	go startCommonProcedures(buttonsChan, obstructChan, stopChan)
	go roleChangeHandler(newRoleChan)

	newRoleChan <- true
	log.Println(consts.Green, "App started", consts.Neutral)

	blocker := make(chan bool, 1)
	<- blocker
}
