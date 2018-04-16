package network

import (
	"consts"
	"math/rand"
	"time"
	//"log"
	"elevator/common"
	"log"
	"net"
	"encoding/json"
)

func MsgListener(receivedRole chan<- consts.Role, conn *net.UDPConn) {
	var typeJson consts.NotificationData
	buffer := make([]byte, 8192)

	for {
		n, err := conn.Read(buffer[0:])
		if err != nil {
			log.Println(consts.Red, "MsgListener > Reading IP failed", consts.Neutral)
			log.Fatal(err)
		}

		if len(buffer) > 0 {
			//log.Println(consts.Cyan, string(buffer), consts.Neutral)
			err2 := json.Unmarshal(buffer[0:n], &typeJson)
			if err2 != nil {
				log.Println(consts.Red, "MsgListener > unmarshal failed", consts.Neutral)
				log.Fatal(err2)
			} else {

				//log.Println(consts.Cyan, "<- received typeJson", typeJson, consts.Neutral)

				switch typeJson.Code {
				case consts.FindRole:
					var role consts.Role
					json.Unmarshal(typeJson.Data, &role)
					receivedRole <- role
				}
			}
		}
	}
}
	// TODO implement role decision
func FindOutRole(ipAddr string) (consts.Role) {
    /**
     * generate random number
     * exchange all numbers between nodes
     * first two bigger numbers are Master and Backup
     */
     //Create broadcast sending address
	//masterConn := GetSendConn(GetBroadcastAddress()+consts.MasterPort)
	//ipAddr = IncreasePortForAddress(masterConn.LocalAddr().String())
	//common.ElevatorState.SetMasterConn(masterConn)
	//conn := GetListenConn(ipAddr)
    a := 0
    ready := false
	rand.Seed(time.Now().UnixNano())
	random := rand.Intn(6) + 2
	log.Print("Random number: ", random)

	var timeout = time.NewTimer(0)
	//receivedRole := make(chan consts.Role)

	//Goroutine for forwarding IP
	//go MsgListener(receivedRole, conn)

	notification := consts.NotificationData{
		Code: consts.NewSlave,
		Data: common.GetRawJSON(ipAddr),
	}

	msg := common.GetNotification(notification)
	for a != random {
		select {
		case <- timeout.C:
			a++
			ready = true
		default:
			if ready{
				if common.ElevatorState.SendToMaster(msg) {
					log.Println(consts.Cyan, "Sending to master", common.ElevatorState.GetMasterConn().RemoteAddr(), consts.Neutral)
				}
				timeout.Reset(2*time.Second)
				ready = false

			} else if common.ElevatorState.GetRole() != consts.DefaultRole {
				return common.ElevatorState.GetRole()
			}
		}
	}
	return consts.Master
}

func FindOutNewMaster()  {

}

func FindOutNewBackup()  {

}
