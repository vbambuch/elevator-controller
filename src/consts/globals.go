package consts

const MinFloor = 0
const MaxFloor = 3
const NumFloors = MaxFloor + 1

const Address = "localhost"
const Port = "15657"


type MotorDirection int

const (
	MotorUP   MotorDirection = 1
	MotorDOWN                = -1
	MotorSTOP                = 0
)

type ButtonType int

const (
	ButtonUP   ButtonType = 0
	ButtonDOWN            = 1
	ButtonCAB             = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

type QueueType int

const (
	CabQueue 	QueueType = 0
	HallQueue 			  = 1
)
