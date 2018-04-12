package common

import (
	"consts"
	"encoding/json"
	"helper"
	"time"
	"log"
	"net"
)

func GetNotification(d interface{}) (consts.Notification) {
	data, err := json.Marshal(d)
	helper.HandleError(err, "JSON error")

	return data
}

func GetRawJSON(d interface{}) (json.RawMessage) {
	data, err := json.Marshal(d)
	helper.HandleError(err, "JSON raw error")

	return data
}

func clearHallOrder(order consts.ButtonEvent) {
	notification := consts.NotificationData{
		Code: consts.ClearHallOrder,
		Data: GetRawJSON(order),
	}

	msg := GetNotification(notification)
	if ElevatorState.sendToMaster(msg) {
		//log.Println(consts.Blue, "-> clear hall order:", order, consts.Neutral)
	}
}

func handleNewOrder(order consts.ButtonEvent, changeOrderChan chan<- bool) (bool) {
	orderInterrupted := false

	if order.Button == consts.ButtonCAB {
		WriteButtonLamp(order.Button, order.Floor, true)
		log.Println(consts.Blue, "Pushed cab call", order, consts.Neutral)
	} else {
		log.Println(consts.Blue, "Pushed hall call", order, consts.Neutral)
	}

	// elevator is going somewhere => sendElevatorToFloor goroutine has been executed
	if ElevatorState.IsMoving() {
		changeOrderChan <- true
		orderInterrupted = true
	}
	ElevatorState.InsertToOrderArray(order)
	//log.Println(consts.Yellow, "Curr cab array:", ElevatorState.GetOrderArray(), consts.Neutral)

	return orderInterrupted
}

func handleReachedDestination(order consts.ButtonEvent)  {
	ElevatorState.SetDirection(consts.MotorSTOP)
	ElevatorState.SetDoorLight(true)
	ElevatorState.DeleteOrder(order)
	ElevatorState.ClearOrderButton(order)

	if order.Button != consts.ButtonCAB {
		clearHallOrder(order) // distribute to all elevators through Master
	}
	log.Println(consts.Blue, "Clear order", order, consts.Neutral)
}


func sendElevatorToFloor(order consts.ButtonEvent, onFloorChan chan<- bool, changeOrderChan <-chan bool, stopMovingChan <-chan bool) {
	direction := consts.MotorUP

	ElevatorState.SetFree(false) // must be there, even if elevator is on floor

	if ElevatorState.GetFloor() > order.Floor {
		direction = consts.MotorDOWN
	} else if ElevatorState.GetFloor() == order.Floor {
		handleReachedDestination(order)
		onFloorChan <- true
		return
	}

	ElevatorState.SetDoorLight(false)
	ElevatorState.SetDirection(direction)

	for {
		select {
		case <-changeOrderChan:
			log.Println(consts.Yellow, "Change order:", order, consts.Neutral)
			return
		case <-stopMovingChan:
			log.Println(consts.Yellow, "Interrupted by stop button", order, consts.Neutral)
			ElevatorState.SetDirection(consts.MotorSTOP)
			return
		default:
			floor := ElevatorState.GetFloor()

			if floor == order.Floor {
				handleReachedDestination(order)
				onFloorChan <- true
				return
			} else {
				//log.Println(consts.Red, "floor:", floor, consts.Neutral)
			}
			time.Sleep(consts.PollRate)
		}
	}
}


func ButtonsHandler(
	localButtonsChan <-chan consts.ButtonEvent,
	remoteHallButtonChan <-chan consts.ButtonEvent,
	obstructChan <-chan bool,
	stopChan <-chan bool)  {

	var timeout = time.NewTimer(0)
	free := false
	orderInterrupted := false
	onFloorChan := make(chan bool)
	changeOrderChan := make(chan bool)
	stopMovingChan := make(chan bool)


	for {
		select {
		case <- onFloorChan:
			timeout.Reset(3 * time.Second)
		case <- timeout.C:
			ElevatorState.SetDoorLight(false)
			ElevatorState.SetFree(true)
			free = true

			// elevator is ready for another hall call
			if ElevatorState.GetHallProcessing() {
				ElevatorState.SetHallProcessing(false)
			}

		case button := <-localButtonsChan:
			if button.Button == consts.ButtonCAB {		// cab button pressed
				if ElevatorState.NewOrder(button) {
					orderInterrupted = handleNewOrder(button, changeOrderChan)
				}
			} else { 									// hall button pressed
				notification := consts.NotificationData{
					Code: consts.SlaveHallOrder,
					Data: GetRawJSON(button),
				}

				msg := GetNotification(notification)
				if ElevatorState.sendToMaster(msg) {
					log.Println(consts.Blue, "-> hall order:", button, consts.Neutral)
				}
			}

		case hallOrder := <-remoteHallButtonChan:
			ElevatorState.SetHallProcessing(true)
			orderInterrupted = handleNewOrder(hallOrder, changeOrderChan)

		case <- obstructChan:
			ElevatorState.SetDirection(consts.MotorSTOP)
			// TODO close all connections and end peacefully
			log.Panic(consts.Red, "OBSTRUCTION! PANIC!", consts.Yellow)

		case stop := <- stopChan:
			log.Printf("stop: %+v\n", stop)
			if stop && ElevatorState.GetStopButton() {
				ElevatorState.SetStopButton(false)
			} else if stop {
				ElevatorState.SetStopButton(true)
				if ElevatorState.IsMoving() || ElevatorState.OrderArrayNotEmpty() {
					stopMovingChan <- true
					orderInterrupted = true
				}
			}

		default:
			if ElevatorState.GetStopButton() == false {
				// just finished previous order or an interrupting cab order
				if ElevatorState.OrderArrayNotEmpty() && (free || orderInterrupted) {
					// get first cab order
					order := ElevatorState.GetOrder()
					log.Println(consts.Blue, "Read from order array", order, consts.Neutral)
					go sendElevatorToFloor(order, onFloorChan, changeOrderChan, stopMovingChan)
					free = false
					orderInterrupted = false
				}
			}
		}
	}
}


func PeriodicNotifications(ipAddr string) {
	for {
		data := consts.PeriodicData{
			ListenIP:       ipAddr,
			Floor:          ElevatorState.GetFloor(),
			Direction:      ElevatorState.GetDirection(),
			OrderArray:     ElevatorState.GetOrderArray(),
			Free:           ElevatorState.GetFree(),
			HallProcessing: ElevatorState.GetHallProcessing(),
		}
		notification := consts.NotificationData{
			Code: consts.SlavePeriodicMsg,
			Data: GetRawJSON(data),
		}

		msg := GetNotification(notification)
		if ElevatorState.sendToMaster(msg) {
			//log.Println(consts.Blue, "-> periodic", *e.orderArray, consts.Neutral)
		}
		//time.Sleep(1 * time.Second)
		time.Sleep(consts.PollRate)
	}
}

func ListenIncomingMsg(receivedHallChan chan<- consts.ButtonEvent, conn *net.UDPConn) {
	var typeJson consts.NotificationData
	buffer := make([]byte, 8192)
	//receivedOrder := make(chan consts.ButtonEvent)

	for {
		n, err := conn.Read(buffer[0:])
		if err != nil {
			log.Println(consts.Blue, "reading slave failed", consts.Neutral)
			log.Fatal(err)
		}
		//log.Println(consts.Blue, buffer, consts.Neutral)
		if len(buffer) > 0 {
			//log.Println(consts.Blue, string(buffer), consts.Neutral)
			err2 := json.Unmarshal(buffer[0:n], &typeJson)
			if err2 != nil {
				log.Println(consts.Blue, "unmarshal slave failed", consts.Neutral)
				log.Fatal(err2)
			} else {

				//log.Println(consts.Blue, "<- received typeJson", typeJson, consts.Neutral)

				switch typeJson.Code {
				case consts.MasterHallOrder:
					order := consts.ButtonEvent{}
					json.Unmarshal(typeJson.Data, &order)
					log.Println(consts.Blue, "<- hall order:", order, consts.Neutral)
					receivedHallChan <- order
				case consts.MasterHallLight:
					order := consts.ButtonEvent{}
					json.Unmarshal(typeJson.Data, &order)
					//log.Println(consts.Blue, "<- hall light:", order, consts.Neutral)
					WriteButtonLamp(order.Button, order.Floor, true)
				case consts.ClearHallOrder:
					order := consts.ButtonEvent{}
					json.Unmarshal(typeJson.Data, &order)
					//log.Println(consts.Blue, "<- clear order:", order, consts.Neutral)
					ElevatorState.ClearOrderButton(order)
				}
			}
		}
	}
}
