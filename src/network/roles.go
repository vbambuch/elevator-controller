package network

import (
	"consts"
	"math/rand"
	"time"
	"elevator/common"
	"log"
)

/**
 * Generate random number of attempts.
 * Try n-times to contact master otherwise become master.
 */
func FindOutRole(ipAddr string) (consts.Role) {
    a := 0
    ready := false
	rand.Seed(time.Now().UnixNano())
	random := rand.Intn(6) + 2
	log.Print("Random number: ", random)

	var timeout = time.NewTimer(0)

	// create packet for master
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
			if ready { // send init packet to master
				if common.ElevatorState.SendToMaster(msg) {
					log.Println(consts.Cyan, "Sending to master", common.ElevatorState.GetMasterConn().RemoteAddr(), consts.Neutral)
				}
				timeout.Reset(2*time.Second)
				ready = false

			} else if common.ElevatorState.GetRole() != consts.DefaultRole { // elevator has received role already 
				return common.ElevatorState.GetRole()
			}
		}
	}
	return consts.Master // number of attempts has been emptied => become master
}
