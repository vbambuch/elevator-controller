package main

import (
	"fmt"
	"log"
	"net"
	"time"
	"consts"
	"strings"
)


/**
 *
 */
func getInstruction(b0, b1, b2, b3 string) ([]byte){
	return []byte(strings.Join([]string{b0, b1, b2, b3}, ""))
}

func handleError(err error, message string) {
	if err != nil {
		fmt.Println(message)
		log.Fatal(err)
	}

}

func sendMessage(socket *net.TCPConn, finished chan bool) {
	for {
		msg := getInstruction(motorDirection, consts.motorUp, consts.emptyByte, consts.emptyByte)
		_, err := socket.Write(msg)
		handleError(err, "write byte")


		time.Sleep(2000 * time.Millisecond)
		fmt.Println("message has been written 2")
	}
	finished <- true
}

func main() {
	finished := make(chan bool, 1)


	// Set up send socket
	sendAddr, err := net.ResolveTCPAddr("tcp", ":15657")
	handleError(err, "resolve tcp")

	socket, err := net.DialTCP("tcp", nil, sendAddr)
	handleError(err, "Dial error")

	// Set up listen socket
	listenAddr, err := net.ResolveTCPAddr("tcp", ":15657")
	handleError(err, "Listen addr error")

	listener, err := net.ListenTCP("tcp", listenAddr)
	handleError(err, "Listener error")

	//fmt.Println("listener ready")

	//msg := getInstruction(floorIndicator, "\xaa", emptyByte, emptyByte)
	//msg := "\\x01\\xfe\\x00\\x00"
	//msg := "\x02\xfe\x00\x00"
	//msg := []byte{1, 254, 0, 0}
	//fmt.Println(string(msg))
	//fmt.Println(msg)
	//fmt.Println(hex.EncodeToString(msg))
	//_, err = socket.Write([]byte(msg))
	//
	// receive message
	//conn, err := listener.AcceptTCP()
	//handleError(err, "Accept error")
	//
	//fmt.Println("connection accepted")

	go sendMessage(socket, finished)

	<- finished
}
