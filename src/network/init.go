package network

import (
	"net"
	"helper"
)

func ReceiveMessage(socket *net.TCPConn, stateChange chan<- []byte) {
	buf := make([]byte, 4)

	for {
		//fmt.Println("Waiting for data...")
		n, err := socket.Read(buf[:])
		helper.HandleError(err, "Read socket error")

		if n == 4{
			stateChange <- buf
		}
		//fmt.Println("Received", n, "bytes:", buf[0], buf[1], buf[2], buf[3])
	}
}

func SendMessage(socket *net.TCPConn, instrChannel <-chan []byte) {
	for {
		msg := <- instrChannel
		_, err := socket.Write(msg)
		helper.HandleError(err, "write byte")

		//fmt.Println("message has been written")
	}
}
