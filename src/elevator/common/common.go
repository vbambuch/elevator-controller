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
		log.Println(consts.Blue, "-> clear order:", order, consts.Neutral)
	}
}

func handleReachedDestination(order consts.ButtonEvent)  {
	ElevatorState.SetDirection(consts.MotorSTOP)
	ElevatorState.SetDoorLight(true)
	clearHallOrder(order)

	if order.Button == consts.ButtonCAB {
		ElevatorState.DeleteOrder(order)
		log.Println(consts.Blue, "Clear cab order", order, consts.Neutral)
	}
}


func sendElevatorToFloor(order consts.ButtonEvent, onFloorChan chan<- bool, interruptCab <-chan bool) {
	direction := consts.MotorUP

	if ElevatorState.GetFloor() > order.Floor {
		direction = consts.MotorDOWN
	} else if ElevatorState.GetFloor() == order.Floor {
		handleReachedDestination(order)
		onFloorChan <- true
		return
	}

	ElevatorState.SetDoorLight(false)
	ElevatorState.SetDirection(direction)
	ElevatorState.SetFree(false)

	for {
		select {
		case <- interruptCab:
			log.Println(consts.Yellow, "Interrupt:", order, consts.Neutral)
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


func OrderHandler(cabButtonChan <-chan consts.ButtonEvent, hallButtonChan <-chan consts.ButtonEvent)  {
	var timeout = time.NewTimer(0)
	free := false
	cabInterrupted := false
	onFloorChan := make(chan bool)
	interruptCab := make(chan bool)

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

		case cabOrder := <- cabButtonChan:
			if !ElevatorState.OrderExists(cabOrder) {
				if ElevatorState.GetDirection() != consts.MotorSTOP {
					interruptCab <- true
					cabInterrupted = true
				}
				log.Println(consts.Blue, "Pushed to cab queue", cabOrder, consts.Neutral)
				ElevatorState.InsertToCabArray(cabOrder)
				log.Println(consts.Yellow, "Curr cab array:", ElevatorState.GetCabArray(), consts.Neutral)
			}

		case hallOrder := <- hallButtonChan:
			ElevatorState.SetHallProcessing(true)
			if free {
				log.Println(consts.Blue, "Free for hall", hallOrder.Floor, consts.Neutral)
				go sendElevatorToFloor(hallOrder, onFloorChan, interruptCab)
				free = false
			} else {
				interruptCab <- true
				log.Println(consts.Blue, "Interrupt and hall", hallOrder.Floor, consts.Neutral)
				go sendElevatorToFloor(hallOrder, onFloorChan, interruptCab)
				free = false
			}

		default:
			if len(ElevatorState.GetCabArray()) != 0 && (free || cabInterrupted) {
				// get first cab order
				queueOrder := ElevatorState.GetCabOrder()
				log.Println(consts.Blue, "Read from cab queue", queueOrder, consts.Neutral)
				go sendElevatorToFloor(queueOrder, onFloorChan, interruptCab)
				free = false
				cabInterrupted = false
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
			CabArray:       ElevatorState.GetCabArray(),
			Free:           ElevatorState.GetFree(),
			HallProcessing: ElevatorState.GetHallProcessing(),
		}
		notification := consts.NotificationData{
			Code: consts.SlavePeriodicMsg,
			Data: GetRawJSON(data),
		}

		msg := GetNotification(notification)
		if ElevatorState.sendToMaster(msg) {
			//log.Println(consts.Blue, "-> periodic", *e.cabArray, consts.Neutral)
		}
		//time.Sleep(1 * time.Second)
		time.Sleep(consts.PollRate)
	}
}

func HallOrderNotifications(sendHallChan <-chan consts.ButtonEvent)  {
	for {
		order := <-sendHallChan
		notification := consts.NotificationData{
			Code: consts.SlaveHallOrder,
			Data: GetRawJSON(order),
		}

		msg := GetNotification(notification)
		if ElevatorState.sendToMaster(msg) {
			log.Println(consts.Blue, "-> hall order:", order, consts.Neutral)
		}
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
					log.Println(consts.Blue, "<- hall light:", order, consts.Neutral)
					WriteButtonLamp(order.Button, order.Floor, true)
				case consts.MasterBroadcastIP:
					var ip string
					json.Unmarshal(typeJson.Data, &ip)
					log.Println(consts.Blue, "<- master ip:", ip, consts.Neutral)
				case consts.ClearHallOrder:
					order := consts.ButtonEvent{}
					json.Unmarshal(typeJson.Data, &order)
					log.Println(consts.Blue, "<- clear order:", order, consts.Neutral)
					ElevatorState.ClearOrderButton(order)
				}
			}
		}
	}
}
