package elevator

import (
	"consts"
	"fmt"
	"time"
	//"helper"
)


func stateHandler(floorChan <-chan int, obstructChan, stopChan <-chan bool, buttonsChan <-chan consts.ButtonEvent, stateChan chan<- Elevator, orderChan chan<- consts.ButtonEvent) {

	for {
		select {
		case button := <-buttonsChan:
			//fmt.Printf("%+v\n", button)
			if button != ElevatorState.GetOrderButton() {
				ElevatorState.SetOrderButton(button)
				orderChan <- button
			}

		case floor := <-floorChan:
			//fmt.Printf("floor: %+v\n", floor)
			if floor != ElevatorState.GetFloor() {
				if floor == consts.MinFloor || floor == consts.MaxFloor {
					ElevatorState.SetDirection(consts.MotorSTOP)
				}
				ElevatorState.SetFloorIndicator(floor)
			}

		case obstruct := <-obstructChan:
			//fmt.Printf("%+v\n", obstruct)
			if obstruct != ElevatorState.GetObstruction() {
				ElevatorState.SetObstruction(obstruct)
			}

		case stop := <-stopChan:
			//fmt.Printf("%+v\n", stop)
			if stop != ElevatorState.GetStopButton() {
				ElevatorState.SetStopButton(stop)

				for f := 0; f < consts.NumFloors; f++ {
					for b := consts.ButtonUP; b < consts.ButtonCAB; b++ {
						WriteButtonLamp(consts.ButtonType(b), f, false)
					}
				}
			}
		}
		stateChan <- ElevatorState
	}
}

func orderHandler(orderChan <-chan consts.ButtonEvent, stateChan <-chan Elevator)  {
	var timeout = time.NewTimer(0)
	ready := false
	onFloorChan := make(chan bool)

	for {
		select {
		case <- onFloorChan:
			timeout.Reset(3 * time.Second)
		case <- timeout.C:
			ElevatorState.SetDoorLight(false)
			ready = true
		case order := <- orderChan:
			if order.Button == consts.ButtonCAB {
				if ready {
					fmt.Printf("Ready for cab %d\n", order.Floor)
					go SendElevatorToFloor(order, stateChan, onFloorChan)
					ready = false
				} else {
					fmt.Printf("Pushed to cab queue %+v\n", order)
					ElevatorState.GetQueue(consts.CabQueue).Push(order)
				}
			} else {
				if ready {
					fmt.Printf("Ready for hall %d\n", order.Floor)
					go SendElevatorToFloor(order, stateChan, onFloorChan)
					ready = false
				} else {
					fmt.Printf("Pushed to hall queue %+v\n", order)
					ElevatorState.GetQueue(consts.HallQueue).Push(order)
				}
			}
		default:
			if ElevatorState.GetQueue(consts.HallQueue).Len() != 0 &&
			   ElevatorState.GetQueue(consts.CabQueue).Len() == 0 &&
				ready {
				// pop order from hall queue
				queueOrder := ElevatorState.GetQueue(consts.HallQueue).Pop().(consts.ButtonEvent)
				fmt.Printf("Pop from hall queue %+v\n", queueOrder)
				go SendElevatorToFloor(queueOrder, stateChan, onFloorChan)
				ready = false
			} else if ElevatorState.GetQueue(consts.CabQueue).Len() != 0 && ready {
				// pop order from cab queue
				queueOrder := ElevatorState.GetQueue(consts.CabQueue).Pop().(consts.ButtonEvent)
				fmt.Printf("Pop from cab queue %+v\n", queueOrder)
				go SendElevatorToFloor(queueOrder, stateChan, onFloorChan)
				ready = false
			}
		}

	}
}

func handleReachedDestination(order consts.ButtonEvent)  {
	ElevatorState.SetDirection(consts.MotorSTOP)
	ElevatorState.SetDoorLight(true)
	ElevatorState.ClearOrderButton(order)
}

func SendElevatorToFloor(order consts.ButtonEvent, stateChan <-chan Elevator, onFloorChan chan<- bool) {
	direction := consts.MotorUP

	if ElevatorState.floor > order.Floor {
		direction = consts.MotorDOWN
	} else if ElevatorState.floor == order.Floor {
		handleReachedDestination(order)
		onFloorChan <- true
		return
	}

	ElevatorState.SetDoorLight(false)
	ElevatorState.SetDirection(direction)

	for {
		info := <-stateChan

		if info.floor == order.Floor {
			handleReachedDestination(order)
			onFloorChan <- true
			return
		}
	}
}

func Init() (chan Elevator) {

	InitIO()

	buttonsChan := make(chan consts.ButtonEvent)
	floorChan := make(chan int)
	obstructChan := make(chan bool)
	stopChan := make(chan bool)
	orderChan := make(chan consts.ButtonEvent)
	stateChan := make(chan Elevator, 4)

	go PollButtons(buttonsChan)
	go PollFloorSensor(floorChan)
	go PollObstructionSwitch(obstructChan)
	go PollStopButton(stopChan)

	go stateHandler(floorChan, obstructChan, stopChan, buttonsChan, stateChan, orderChan)
	go orderHandler(orderChan, stateChan)

	// wait for initialization of elevator
	for ElevatorState.GetFloor() == -1 {
		ElevatorState.SetDirection(consts.MotorUP)
		time.Sleep(pollRate)
	}
	ElevatorState.SetDirection(consts.MotorSTOP)
	return stateChan
}
