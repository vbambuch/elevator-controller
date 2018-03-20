package network

import (
	"consts"
	"net"
	"helper"
)

// TODO implement role decision
func FindOutRole() (consts.Role) {
    /**
     * generate random number
     * exchange all numbers between nodes
     * first two bigger numbers are Master and Backup
     */
	return consts.Master
}

func FindOutNewMaster()  {

}

func FindOutNewBackup()  {

}

// TODO implement
func GetMasterConn() (*net.UDPConn) {
	addr, err := net.ResolveUDPAddr("udp", "localhost:12345")
	helper.HandleError(err, "UDP resolve failed")

	conn, err := net.DialUDP("udp", nil, addr)
	helper.HandleError(err, "UDP dial failed")

	return conn
}

func GetListenConn() (*net.UDPConn) {
	addr, err := net.ResolveUDPAddr("udp", "localhost:12345")
	helper.HandleError(err, "UDP resolve failed")

	conn, err := net.ListenUDP("udp", addr)
	helper.HandleError(err, "UDP listen failed")

	return conn
}


// TODO delete
func GetSlaveTestSendConn() (*net.UDPConn) {

	addr, err := net.ResolveUDPAddr("udp", "localhost:50000")
	helper.HandleError(err, "UDP resolve failed")

	conn, err := net.DialUDP("udp", nil, addr)
	helper.HandleError(err, "UDP dial failed")

	return conn
}

func GetSlaveTestListenConn() (*net.UDPConn) {

	addr, err := net.ResolveUDPAddr("udp", "localhost:50000")
	helper.HandleError(err, "UDP resolve failed")

	conn, err := net.ListenUDP("udp", addr)
	helper.HandleError(err, "UDP dial failed")

	return conn
}
