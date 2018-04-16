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

func splitAddress(address string) (string, int) {
	re, _ := regexp.Compile("(.*):(.*)")
	match := re.FindStringSubmatch(address)
	port, _ := strconv.Atoi(match[2])

	return match[1], port
}

func IncreasePortForAddress(addr string) string {
	localAddr, localPort := splitAddress(addr)

	return localAddr+":"+strconv.Itoa(localPort+1)
}

func getPrivateAddress() string {
	conn := GetSendConn("8.8.8.8:80")
	if conn != nil {
		addr, _ := splitAddress(conn.LocalAddr().String())
		conn.Close()
		return addr
	}
	return ""
}

func replaceLastOctet(addr string, octet string) string {
	re, _ := regexp.Compile("([0-9]+)$")
	return re.ReplaceAllString(addr, octet)
}

func GetBroadcastAddress() string {
	return replaceLastOctet(getPrivateAddress(), "255")+":"
}

func GetNetworkAddress() string {
	return replaceLastOctet(getPrivateAddress(), "0")+":"
}
