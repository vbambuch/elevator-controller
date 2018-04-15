package network

import (
	"net"
	"helper"
	"regexp"
	"strconv"
)

func GetSendConn(ipAddr string) (*net.UDPConn) {
	//log.Println(consts.Grey, "Send conn ip:", ipAddr, consts.Neutral)

	addr, err := net.ResolveUDPAddr("udp", ipAddr)
	helper.HandleError(err, "UDP resolve failed")

	conn, err := net.DialUDP("udp", nil, addr)
	helper.HandleError(err, "UDP dial failed")

	return conn
}


func GetListenConn(ipAddr string) (*net.UDPConn) {
	//log.Println(consts.Grey, "Listen conn ip:", ipAddr, consts.Neutral)

	addr, err := net.ResolveUDPAddr("udp", ipAddr)
	helper.HandleError(err, "UDP resolve failed")

	conn, err := net.ListenUDP("udp", addr)
	helper.HandleError(err, "UDP listen failed")

	return conn
}

func IncreasePortForAddress(addr string) string {
	re := regexp.MustCompile("(.*):(.*)")
	match := re.FindStringSubmatch(addr)

	localAddr:= match[1]
	tmpPort, _ := strconv.Atoi(match[2])
	localPort := tmpPort + 1

	return localAddr+":"+strconv.Itoa(localPort)
}
