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


func startCommonProcedures(receivedCabChan <-chan consts.ButtonEvent, sendHallChan <-chan consts.ButtonEvent)  {
	receivedHallChan := make(chan consts.ButtonEvent)

	ipAddr := "localhost:"+consts.MyPort
	conn := network.GetSlaveListenConn(ipAddr)

	go common.ElevatorState.PeriodicNotifications(ipAddr)
	go common.ElevatorState.HallOrderNotifications(sendHallChan)
	go common.ElevatorState.ListenIncomingMsg(receivedHallChan, conn)
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
		log.Println(consts.Green, "Setting master connection...", consts.Neutral)
		masterConn := network.GetMasterSendConn()

		myIP := masterConn.LocalAddr()
		log.Println(consts.Green, "My ListenIP:", myIP.String(), consts.Neutral)

		common.ElevatorState.SetMasterConn(masterConn)

		switch common.ElevatorState.GetRole() {
		case consts.Master:
			log.Println(consts.Blue, "It's master", consts.Neutral)
			go master.StartMaster()
		case consts.Backup:
			log.Println(consts.Blue, "It's backup", consts.Neutral)
			go slave.StartBackup(orderChan, masterConn)
		case consts.Slave:
			log.Println(consts.Blue, "It's slave", consts.Neutral)
			//go slave.StartSlave(orderChan, masterConn)
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

	// TODO remove when network is done...
	masterPort := flag.String("masterPort", "20002", "master localhost port")
	myPort := flag.String("myPort", "20000", "my localhost port")
	elPort := flag.String("elPort", "20001", "my elevator port")
	myRole := flag.Int("myRole", 1, "1: Master, 2: Backup, 3: Slave")
	flag.Parse()

	//log.Println(consts.Green, "Master port:", *masterPort, consts.Neutral)
	//log.Println(consts.Green, "My port:", *myPort, consts.Neutral)

	//log.Println(consts.Green, "My elevator port:", *elPort, consts.Neutral)
	//log.Println(consts.Green, "My role:", *myRole, consts.Neutral)

	consts.ElevatorPort = *elPort
	consts.MasterPort = *masterPort
	consts.MyPort = *myPort

	// TODO ...remove when network is done

	errorChan := make(chan consts.ElevatorError)
	newRoleChan := make(chan bool)
	//stateChan := make(chan common.Elevator)
	orderChan := make(chan consts.ButtonEvent)

	// master-backup-slave decision
	//roleDecision()	// TODO uncomment
	common.ElevatorState.SetRole(consts.Role(*myRole)) // TODO remove

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
