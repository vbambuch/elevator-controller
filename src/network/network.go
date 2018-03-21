package network

import (

	"consts"
	"fmt"
	"log"
	"time"
	"encoding/json"
)



//Function for pinging other elevators to check if they are alive
func pingElevators(outgoingMsg chan<- consts.Message){
	const pingInterval = 1000 * time.Millisecond
	alive := consts.Message{Category: consts.Alive, Addr: consts.Laddr, Floor: -1, Button: -1}
	for {
		outgoingMsg <- alive
		time.Sleep(pingInterval)
	}
}


//Function for broadcasting the IP address of the Master PC
func masterNotify(outgoingMsg chan<- consts.Message){
	const notifyInterval = 2000 * time.Millisecond
	master := consts.Message{Category: consts.Alive, Addr: consts.Laddr, Master:true, Floor: -1, Button: -1}
	for {
		outgoingMsg <- master
		time.Sleep(notifyInterval)
	}
}

//Function for broadcasting the IP address of the Slave PC
func slaveNotify(outgoingMsg chan<- consts.Message){
	const notifyInterval = 2000 * time.Millisecond
	slave := consts.Message{Category: consts.Alive, Addr: consts.Laddr, Master:false, Floor: -1, Button: -1}
	for {
		outgoingMsg <- slave
		time.Sleep(notifyInterval)
	}
}



//Function for sending messages, constantly checks for messages to send on the network
func sendMessage(outgoingMsg <-chan consts.Message, UDPSend chan<- UDPMessage){

	for {
		message := <-outgoingMsg

		//Marshal JSON message
		json_message, err := json.Marshal(message)
		if err != nil {
			log.Printf("%s JSON Marshal error: %v\n%s", consts.Red, err, consts.Neutral)
		}

		UDPSend <- UDPMessage{raddr: "broadcast", data: json_message, length: len(json_message)}
	}

}

//Function for receiving messages, constantly checks for incoming messages on the network
func receiveMessage(incomingMsg chan<- consts.Message, udpReceive <-chan UDPMessage){

	for {
		UDPMessage := <-udpReceive
		var message consts.Message

		//Unmarshal JSON message
		if err := json.Unmarshal(UDPMessage.data[:UDPMessage.length], &message); err != nil {
			fmt.Printf("JSON Unmarshal error: %s\n", err)
		}

		message.Addr = UDPMessage.raddr
		incomingMsg <- message
	}
}


//Main function starting the network on one PC
func Initialize(outgoingMsg, incomingMsg chan consts.Message){

	// Ports and message size
	const localListenPort = 20008
	const broadcastListenPort = 20009
	const messageSize = 1024

	//Channels for sending and receiving
	var send = make(chan UDPMessage)
	var receive = make(chan UDPMessage, 10)

	//Initialize UDP connections
	err := UDPInit(localListenPort, broadcastListenPort, messageSize, send, receive)
	if err != nil {
		fmt.Print("UDPInit() error: %v \n", err)
	}

	//Goroutines
	// 1. ping elevators
	// 2. send messages
	// 3. receive messages
	go pingElevators(outgoingMsg)
	go sendMessage(outgoingMsg, send)
	go receiveMessage(incomingMsg, receive)

	//Master and slave notification
	//go masterNotify(outgoingMsg)
	//go slaveNotify(outgoingMsg)

	time.Sleep(100*time.Millisecond)
	log.Println(consts.Green, "Network initialized.", consts.Neutral)
}