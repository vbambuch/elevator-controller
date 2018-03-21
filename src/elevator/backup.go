package elevator

import (
	"net"
	"consts"
)

// Backup
/**
 * defer old instance
 * create Backup
 * listen for incoming DB syncs from Master
 * ping Master
 * -> became Master if prev Master failed
 * do same things as Slave
 */
func StartBackup(orderChan <-chan consts.ButtonEvent, masterConn *net.UDPConn) {

}
