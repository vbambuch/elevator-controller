package network

import (
	"net"
	"helper"
	"strings"
	//"fmt"
)

func MessageReceiver(socket *net.TCPConn, stateChange chan<- []byte) {
	buf := make([]byte, 4)

	for {
		n, err := socket.Read(buf[:])
		helper.HandleError(err, "Read socket error")

		if n == 4{
			stateChange <- buf
		}
		//fmt.Println("Received", n, "bytes:", buf[0], buf[1], buf[2], buf[3])
	}
}

func MessageSender(socket *net.TCPConn, instrChannel <-chan []byte) {
	for {
		msg := <- instrChannel
		_, err := socket.Write(msg)
		helper.HandleError(err, "write byte")
	}
}

func GetSocket(addr string, port string) (*net.TCPConn) {
	address := strings.Join([]string{addr, port}, ":")

	// Set up send socket
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	helper.HandleError(err, "Resolve tcp")

	socket, err := net.DialTCP("tcp", nil, tcpAddr)
	helper.HandleError(err, "Dial tcp")

	return socket
}
