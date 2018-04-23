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
	common.ElevatorState.SetRole(role)
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

func resolveRoleChange(masterFailed chan<- consts.BackupSync, backupData consts.BackupSync) {
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

/**
 * There are two ways how to change the role.
 * - new backup is elected or master by default
 * - master fails, backup will become new master
 */
func roleChangeHandler(newRoleChan <-chan bool)  {
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


func main() {
	// define available script arguments
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

	// create master connection and define own listening address
	masterConn := network.GetSendConn(network.GetBroadcastAddress()+consts.MasterPort)
	common.ElevatorState.SetMasterConn(masterConn)
	ipAddr := network.IncreasePortForAddress(masterConn.LocalAddr().String())
	log.Println(consts.Green, "My IP:", ipAddr, consts.Neutral)

	// init elevator driver
	newRoleChan := make(chan bool)
	buttonsChan, obstructChan, stopChan := common.Init()

	// start error detection
	go startCommonProcedures(buttonsChan, obstructChan, stopChan, newRoleChan, ipAddr)
	go roleChangeHandler(newRoleChan)

	// master-backup-slave decision
	time.Sleep(500 * time.Millisecond)
	roleDecision(ipAddr)

	// quick and dirty fix
	// slave is by default, backup is election in different way:
	// common.ListenIncomingMsg => FindRole case
	if common.ElevatorState.GetRole() == consts.Master {
		newRoleChan <- true
	}
	log.Println(consts.Green, "App started", consts.Neutral)

	blocker := make(chan bool, 1)
	<- blocker
}
