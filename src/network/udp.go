package network

import (

	"consts"
	"time"
	"net"
	"strconv"
	"fmt"
	"log"
	"strings"
)

//The struct for the UDP message
type UDPMessage struct{
	data []byte
	length int
	raddr string
}

//UDP connection struct
type UDPConn struct {
	addr string
	timer *time.Timer
}

//Broadcast address
var baddr *net.UDPAddr


//Function for making the UDP sending server
func sendServer(lconn, bconn *net.UDPConn, send_ch <-chan UDPMessage) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error in send server: %s \n Closing connection.", r)
			lconn.Close()
			bconn.Close()
		}
	}()

	var err error
	var n int

	for {
		//fmt.Printf("UDP send server: Waiting on new value on global sending channel \n")

		msg := <-send_ch

		//fmt.Printf("UDP send server: Writing %s \n", msg.data)
		if msg.raddr == "broadcast" {
			n, err = lconn.WriteToUDP(msg.data, baddr)
		} else {
			raddr, err := net.ResolveUDPAddr("udp", msg.raddr)
			if err != nil {
				fmt.Printf("Error in UDP send server: could not resolve raddr\n")
				panic(err)
			}
			n, err = lconn.WriteToUDP(msg.data, raddr)
		}
		if err != nil || n < 0 {
			fmt.Printf("Error in UDP send server: writing\n")
			panic(err)
		}
		//fmt.Printf("UDP sending server: Sent %s to %s \n", msg.data, msg.raddr)
	}
}

//Function for making the UDP receiving server
func receiveServer(lconn, bconn *net.UDPConn, message_size int, receive_ch chan<- UDPMessage) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error in UDP receive server: %s \n Closing connection.", r)
			lconn.Close()
			bconn.Close()
		}
	}()

	bconn_rcv_ch := make(chan UDPMessage)
	lconn_rcv_ch := make(chan UDPMessage)

	go readConn(lconn, message_size, lconn_rcv_ch)
	go readConn(bconn, message_size, bconn_rcv_ch)

	for {
		select {

		case buf := <-bconn_rcv_ch:
			receive_ch <- buf

		case buf := <-lconn_rcv_ch:
			receive_ch <- buf
		}
	}
}

//Reading a connection
func readConn(conn *net.UDPConn, message_size int, rcv_ch chan<- UDPMessage) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error in UDP connection reader: %s \n Closing connection.", r)
			conn.Close()
		}
	}()

	for {
		buf := make([]byte, message_size)
		//fmt.Printf("UDP connection reader: Waiting on data from UDPConn\n")
		n, raddr, err := conn.ReadFromUDP(buf)
		//fmt.Printf("UDP connection reader: Received %s from %s \n", string(buf), raddr.String())
		if err != nil || n < 0 {
			fmt.Printf("Error: UDP connection reader: reading\n")
			panic(err)
		}
		rcv_ch <- UDPMessage{raddr: raddr.String(), data: buf, length: n}
	}
}

//Function for closing connection
func closeConn(lconn, bconn *net.UDPConn) {
	<-consts.CloseConnChan
	lconn.Close()
	bconn.Close()
}


//Main function for initializing the UDP connections
func UDPInit(localListenPort, broadcastListenPort, message_size int, send_ch, receive_ch chan UDPMessage) (err error){

	//Generating broadcast address
	baddr, err = net.ResolveUDPAddr("udp4", "255.255.255.255:"+strconv.Itoa(broadcastListenPort))
	if err != nil {
		return err
	}

	//Generating local address
	tempConn, err := net.DialUDP("udp4", nil, baddr)
	defer tempConn.Close()
	tempAddr := tempConn.LocalAddr()
	laddr, err := net.ResolveUDPAddr("udp4", tempAddr.String())
	laddr.Port = localListenPort
	consts.Laddr = laddr.String()

	//Creating local listening connections
	localListenConn, err := net.ListenUDP("udp4", laddr)
	if err != nil {
		return err
	}

	//Creating listener on broadcast connection
	broadcastListenConn, err := net.ListenUDP("udp", baddr)
	if err != nil {
		localListenConn.Close()
		return err
	}

	//Goroutines
	go receiveServer(localListenConn, broadcastListenConn, message_size, receive_ch)
	go sendServer(localListenConn, broadcastListenConn, send_ch)
	go closeConn(localListenConn, broadcastListenConn)

	log.Println(consts.Green,"\n","Generating the local address...\n", "Network connection: ", strings.ToUpper(laddr.Network()),"\n", "Local address: ", laddr.String(), consts.Neutral)
	log.Println(consts.Green,"\n","Generating the broadcast address...\n", "Network connection: ", strings.ToUpper(baddr.Network()),"\n", "Broadcast address: ", baddr.String(), consts.Neutral)
	return err
}