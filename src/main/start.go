package main

import (
	"network"
	"consts"
	"log"
	"elevator/master"
	"elevator/common"
	"elevator/slave"
	"flag"
	"time"
)

func roleDecision(idAddr string)  {
	common.ElevatorState.SetRole(consts.DefaultRole)

	role := network.FindOutRole(idAddr)
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
	stopChan <-chan bool,
	newRoleChan chan<- bool,
	ipAddr string)  {

	receivedHallChan := make(chan consts.ButtonEvent)
	conn := network.GetListenConn(ipAddr)

	go common.PeriodicNotifications(ipAddr)
	go common.ListenIncomingMsg(receivedHallChan, conn, newRoleChan)
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
			//role := common.ElevatorState.GetRole()
			resolveRoleChange(masterFailed, backupData)
		case backupData = <- masterFailed:
			log.Println(consts.Yellow, "backup data:", backupData, consts.Neutral)
			common.ElevatorState.SetRole(consts.Master)
			resolveRoleChange(masterFailed, backupData)
		}
	}
}


func main() {
	numFloors := flag.Int("numFloors", 4, "elevator id")
	id := flag.Int("id", 0, "elevator id")
	elPort := flag.String("elPort", "15657", "my elevator port")

	flag.Parse()

	log.Println(consts.Green, "Elevator ID:", *id, consts.Neutral)
	log.Println(consts.Green, "Number of floors:", *numFloors, consts.Neutral)
	log.Println(consts.Green, "Elevator port:", *elPort, consts.Neutral)

	consts.ElevatorPort = *elPort
	consts.NumFloors = *numFloors
	consts.MaxFloor = *numFloors - 1

	masterConn := network.GetSendConn(network.GetBroadcastAddress()+consts.MasterPort)
	common.ElevatorState.SetMasterConn(masterConn)
	ipAddr := network.IncreasePortForAddress(masterConn.LocalAddr().String())

	log.Println(consts.Green, "My IP:", ipAddr, consts.Neutral)

	//errorChan := make(chan consts.ElevatorError)
	newRoleChan := make(chan bool)

	//common.ElevatorState.SetRole(consts.Role(*myRole)) // TODO remove

	buttonsChan, obstructChan, stopChan := common.Init()
	//network.Initialize(outgoingMsg, incomingMsg)

	// start error detection
	//go network.ErrorDetection(errorChan)
	//go errorHandler(errorChan, newRoleChan)
	go startCommonProcedures(buttonsChan, obstructChan, stopChan, newRoleChan, ipAddr)
	go roleChangeHandler(newRoleChan)


	time.Sleep(500 * time.Millisecond)
	// master-backup-slave decision
	roleDecision(ipAddr)	// TODO uncomment

	if common.ElevatorState.GetRole() == consts.Master {
		newRoleChan <- true
	}
	log.Println(consts.Green, "App started", consts.Neutral)

	blocker := make(chan bool, 1)
	<- blocker
}
