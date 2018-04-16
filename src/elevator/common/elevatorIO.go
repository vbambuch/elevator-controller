package common

import (
	"time"
	"sync"
	"net"
	"fmt"
	"consts"
	"log"
	"network"
)

var initialized = false
var mutex sync.Mutex
var connect *net.TCPConn

func ReInitIO()  {
	if connect != nil {
		connect.Close()
		connect = nil
	}

	address := consts.LocalAddress+consts.ElevatorPort
	var limit int
	var err error

	for connect == nil {
		mutex.Lock()
		connect, err = network.GetTCPSendConn(address)
		mutex.Unlock()
		if limit == 5 && err != nil {
			log.Fatal(consts.Red, "Elevator connection failed:", err.Error(), consts.Neutral)
		} else if err != nil {
			log.Println(consts.Red, "Elevator connection failed...turn on the elevator.", consts.Neutral)
		} else {
			break
		}
		limit++
		time.Sleep(2 * time.Second)
	}
	log.Println(consts.Green, "Elevator connection reestablished.", consts.Neutral)
	log.Println(consts.Green, ElevatorState.GetPrevFloor(), ElevatorState.GetFloor(), consts.Neutral)
}

func InitIO() {
	if initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	address := consts.LocalAddress+consts.ElevatorPort
	mutex = sync.Mutex{}
	var limit int
	var err error

	for connect == nil {
		connect, err = network.GetTCPSendConn(address)
		if limit == 5 && err != nil {
			log.Fatal(consts.Red, "Elevator connection failed:", err.Error(), consts.Neutral)
		} else if err != nil {
			log.Println(consts.Red, "Elevator connection failed...turn on the elevator.", consts.Neutral)
		}
		limit++
		time.Sleep(2 * time.Second)
	}
	log.Println(consts.Green, "Elevator connection initialized.", consts.Neutral)
	initialized = true
}



func WriteMotorDirection(dir consts.MotorDirection) {
	mutex.Lock()
	defer mutex.Unlock()
	if connect != nil {
		connect.Write([]byte{1, byte(dir), 0, 0})
	}
}

func WriteButtonLamp(button consts.ButtonType, floor int, value bool) {
	mutex.Lock()
	defer mutex.Unlock()
	if connect != nil {
		connect.Write([]byte{2, byte(button), byte(floor), toByte(value)})
	}
}

func WriteFloorIndicator(floor int) {
	mutex.Lock()
	defer mutex.Unlock()
	if connect != nil {
		connect.Write([]byte{3, byte(floor), 0, 0})
	}
}

func WriteDoorOpenLamp(value bool) {
	mutex.Lock()
	defer mutex.Unlock()
	if connect != nil {
		connect.Write([]byte{4, toByte(value), 0, 0})
	}
}

func WriteStopLamp(value bool) {
	mutex.Lock()
	defer mutex.Unlock()
	if connect != nil {
		connect.Write([]byte{5, toByte(value), 0, 0})
	}
}



func pollButtons(receiver chan<- consts.ButtonEvent) {
	prev := make([][3]bool, consts.NumFloors)
	for {
		time.Sleep(consts.PollRate)
		//fmt.Println("Poll buttons")
		for f := 0; f < consts.NumFloors; f++ {
			for b := consts.ButtonType(0); b < 3; b++ {
				v := readButton(b, f)
				if v != prev[f][b] && v != false {
					receiver <- consts.ButtonEvent{f, consts.ButtonType(b)}
				}
				prev[f][b] = v
			}
		}
	}
}

func pollFloorSensor(receiver chan<- int) {
	prev := consts.DefaultValue
	for {
		time.Sleep(consts.PollRate)
		//fmt.Println("Poll sensor")
		v := ReadFloor()
		//log.Println(consts.Grey, "Floor val:", v, consts.Neutral)
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func pollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(consts.PollRate)
		//fmt.Println("Poll stop button")
		v := readStopButton()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func pollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(consts.PollRate)
		//fmt.Println("Poll obstruction")
		v := readObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}



func readButton(button consts.ButtonType, floor int) bool {
	mutex.Lock()
	defer mutex.Unlock()
	if connect != nil {
		connect.Write([]byte{6, byte(button), byte(floor), 0})
	}
	if connect != nil {
		var buf [4]byte
		connect.Read(buf[:])
		return toBool(buf[1])
	}
	return false
}

func ReadFloor() int {
	mutex.Lock()
	defer mutex.Unlock()
	if connect != nil {
		connect.Write([]byte{7, 0, 0, 0})
	}
	if connect != nil {
		var buf [4]byte
		n, _ := connect.Read(buf[:])
		if buf[1] != 0 {
			return int(buf[2])
		} else if n != 0 {
			return consts.MiddleFloor
		} else {
			return consts.ElevatorFailed
		}
	}
	return consts.DefaultValue
}

func readStopButton() bool {
	mutex.Lock()
	defer mutex.Unlock()
	if connect != nil {
		connect.Write([]byte{8, 0, 0, 0})
	}
	if connect != nil {
		var buf [4]byte
		connect.Read(buf[:])
		return toBool(buf[1])
	}
	return false
}

func readObstruction() bool {
	mutex.Lock()
	defer mutex.Unlock()
	if connect != nil {
		connect.Write([]byte{9, 0, 0, 0})
	}
	if connect != nil {
		var buf [4]byte
		connect.Read(buf[:])
		return toBool(buf[1])
	}
	return false
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b = false
	if a != 0 {
		b = true
	}
	return b
}
