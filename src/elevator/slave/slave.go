package slave

import (
	"net"
	"sync"
	"consts"
	"elevator/common"
)

var SlaveSingleton = Slave{}

type Slave struct {
	masterConn net.UDPConn
	mutex      sync.Mutex
}

/**
 * defer old instance
 * create Slave
 * send periodic notifications
 * send orders to Master
 * receive requests from Master
 */
func StartSlave(orderChan <-chan consts.ButtonEvent, masterConn *net.UDPConn) {
	common.ElevatorState.SetMasterConn(masterConn)

	go common.ElevatorState.PeriodicNotifications()
}


