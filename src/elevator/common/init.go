package common

import (
	"consts"
	"time"
	//"helper"
	"log"
)


//func stateHandler(floorChan <-chan int, obstructChan, stopChan <-chan bool, buttonsChan <-chan consts.ButtonEvent, stateChan chan<- Elevator, orderChan chan<- consts.ButtonEvent) {
func stateHandler(floorChan <-chan int, obstructChan, stopChan <-chan bool, buttonsChan <-chan consts.ButtonEvent, cabButtonChan chan<- consts.ButtonEvent, hallButtonChan chan<- consts.ButtonEvent) {
	changed := false


	for {
		select {
		case button := <-buttonsChan:
			//log.Printf("%+v\n", button)
			if button.Button == consts.ButtonCAB {
				cabButtonChan <- button
				WriteButtonLamp(button.Button, button.Floor, true)
			} else {
				//ElevatorState.SetHallButton(button)
				hallButtonChan <- button
			}

			//if button != ElevatorState.GetOrderButton() {
			//	ElevatorState.SetHallButton(button) // prevent order button spam
			//	orderChan <- button
			//	changed = true
			//}

		case floor := <-floorChan:
			//log.Printf("floor: %+v\n", floor)
			if floor != ElevatorState.GetFloor() {
				if floor == consts.MinFloor || floor == consts.MaxFloor &&
					ElevatorState.GetDirection() != consts.MotorSTOP {
					ElevatorState.SetDirection(consts.MotorSTOP)
					ElevatorState.SetFloorIndicator(floor)
				} else if floor == consts.MiddleFloor {
					ElevatorState.SetMiddleFloor(true)
				} else {
					ElevatorState.SetMiddleFloor(false)
					ElevatorState.SetFloorIndicator(floor)
				}
				changed = true
			}

		case obstruct := <-obstructChan:
			log.Printf("%+v\n", obstruct)
			if obstruct != ElevatorState.GetObstruction() {
				ElevatorState.SetObstruction(obstruct)
				changed = true
			}

		case stop := <-stopChan:
			log.Printf("%+v\n", stop)
			if stop != ElevatorState.GetStopButton() {
				ElevatorState.SetStopButton(stop)

				for f := 0; f < consts.NumFloors; f++ {
					for b := consts.ButtonUP; b < consts.ButtonCAB; b++ {
						WriteButtonLamp(consts.ButtonType(b), f, false)
					}
				}
				changed = true
			}
		}
		if changed {
			//fmt.Println("Changed")

			//stateChan <- ElevatorState
			changed = false
		}
	}
}

func handleReachedDestination(order consts.ButtonEvent)  {
	ElevatorState.SetDirection(consts.MotorSTOP)
	ElevatorState.SetDoorLight(true)
	ElevatorState.ClearOrderButton(order) // TODO send to master to clear in all elevators

	if order.Button == consts.ButtonCAB {
		ElevatorState.DeleteOrder(order)
		log.Println(consts.Blue, "Clear cab order", order, consts.Neutral)
	}
}

func SendElevatorToFloor(order consts.ButtonEvent, onFloorChan chan<- bool, interruptCab <-chan bool) {
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

//func Init(stateChan chan<- Elevator, orderChan chan<- consts.ButtonEvent) {
func Init() (chan consts.ButtonEvent, chan consts.ButtonEvent) {

	InitIO()

	buttonsChan := make(chan consts.ButtonEvent)
	floorChan := make(chan int)
	obstructChan := make(chan bool)
	stopChan := make(chan bool)

	cabButtonChan := make(chan consts.ButtonEvent)
	hallButtonChan := make(chan consts.ButtonEvent)

	go PollFloorSensor(floorChan)
	go PollObstructionSwitch(obstructChan)
	go PollStopButton(stopChan)
	go PollButtons(buttonsChan)

	//go stateHandler(floorChan, obstructChan, stopChan, buttonsChan, stateChan, orderChan)
	go stateHandler(floorChan, obstructChan, stopChan, buttonsChan, cabButtonChan, hallButtonChan)

	// wait for initialization of elevator
	setup := true
	time.Sleep(2 * consts.PollRate) // wait for message exchange
	for ElevatorState.GetFloor() == consts.MiddleFloor {
		if setup {
			ElevatorState.SetDirection(consts.MotorUP)
			log.Println(consts.Green, "Elevator is moving to floor...", consts.Neutral)
			setup = false
		}
		time.Sleep(consts.PollRate)
	}
	ElevatorState.SetDirection(consts.MotorSTOP)
	return cabButtonChan, hallButtonChan
}
