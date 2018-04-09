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
func GetMasterSendConn() (*net.UDPConn) {
	addr, err := net.ResolveUDPAddr("udp", "localhost:"+consts.MasterPort)
	helper.HandleError(err, "UDP master resolve failed")

	conn, err := net.DialUDP("udp", nil, addr)
	helper.HandleError(err, "UDP master dial failed")

	return conn
}

func GetMasterListenConn() (*net.UDPConn) {
	addr, err := net.ResolveUDPAddr("udp", "localhost:"+consts.MasterPort)
	helper.HandleError(err, "UDP master resolve failed")

	conn, err := net.ListenUDP("udp", addr)
	helper.HandleError(err, "UDP master listen failed")

	return conn
}

func GetSlaveSendConn(ipAddr string) (*net.UDPConn) {

	addr, err := net.ResolveUDPAddr("udp", ipAddr)
	helper.HandleError(err, "UDP slave resolve failed")

	conn, err := net.DialUDP("udp", nil, addr)
	helper.HandleError(err, "UDP slave dial failed")

	return conn
}


func GetSlaveListenConn(ipAddr string) (*net.UDPConn) {

	addr, err := net.ResolveUDPAddr("udp", ipAddr)
	helper.HandleError(err, "UDP slave resolve failed")

	conn, err := net.ListenUDP("udp", addr)
	helper.HandleError(err, "UDP slave listen failed")

	return conn
}
