package common

import (
	"time"
	"sync"
	"net"
	"fmt"
	"consts"
)

var initialized = false
var mutex sync.Mutex
var connect net.Conn

// reinit IO after obstruction button
func ReInitIO()  {
	if connect != nil {
		connect.Close()
	}
	var err error
	connect, err = net.Dial("tcp", consts.LocalAddress+consts.ElevatorPort)
	if err != nil {
		panic(err.Error())
	}
	initialized = true
}

func InitIO() {
	if initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	address := consts.LocalAddress+consts.ElevatorPort
	mutex = sync.Mutex{}
	var err error
	connect, err = net.Dial("tcp", address)
	if err != nil {
		panic(err.Error())
	}
	initialized = true
}



func WriteMotorDirection(dir consts.MotorDirection) {
	mutex.Lock()
	defer mutex.Unlock()
	connect.Write([]byte{1, byte(dir), 0, 0})
}

func WriteButtonLamp(button consts.ButtonType, floor int, value bool) {
	mutex.Lock()
	defer mutex.Unlock()
	connect.Write([]byte{2, byte(button), byte(floor), toByte(value)})
}

func WriteFloorIndicator(floor int) {
	mutex.Lock()
	defer mutex.Unlock()
	connect.Write([]byte{3, byte(floor), 0, 0})
}

func WriteDoorOpenLamp(value bool) {
	mutex.Lock()
	defer mutex.Unlock()
	connect.Write([]byte{4, toByte(value), 0, 0})
}

func WriteStopLamp(value bool) {
	mutex.Lock()
	defer mutex.Unlock()
	connect.Write([]byte{5, toByte(value), 0, 0})
}



func PollButtons(receiver chan<- consts.ButtonEvent) {
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

func PollFloorSensor(receiver chan<- int) {
	prev := consts.DefaultValue
	for {
		time.Sleep(consts.PollRate)
		//fmt.Println("Poll sensor")
		v := ReadFloor()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func PollStopButton(receiver chan<- bool) {
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

func PollObstructionSwitch(receiver chan<- bool) {
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
	connect.Write([]byte{6, byte(button), byte(floor), 0})
	var buf [4]byte
	connect.Read(buf[:])
	return toBool(buf[1])
}

func ReadFloor() int {
	mutex.Lock()
	defer mutex.Unlock()
	connect.Write([]byte{7, 0, 0, 0})
	var buf [4]byte
	connect.Read(buf[:])
	if buf[1] != 0 {
		return int(buf[2])
	} else {
		return consts.MiddleFloor
	}
}

func readStopButton() bool {
	mutex.Lock()
	defer mutex.Unlock()
	connect.Write([]byte{8, 0, 0, 0})
	var buf [4]byte
	connect.Read(buf[:])
	return toBool(buf[1])
}

func readObstruction() bool {
	mutex.Lock()
	defer mutex.Unlock()
	connect.Write([]byte{9, 0, 0, 0})
	var buf [4]byte
	connect.Read(buf[:])
	return toBool(buf[1])
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
