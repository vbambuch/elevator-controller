package consts

const MinFloor = 0
const MaxFloor = 3
const NumFloors = MaxFloor + 1

const Address = "localhost"
const Port = "15657"

// Elevator consts
const DefaultValue = -1
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

type Role int

const (
	Master	Role = 1
	Backup		 = 0
	Slave 		 = -1
)

// Error detection
type ErrorCode int

const (
	MasterFailed	ErrorCode = 1
	BackupFailed			  = 2
	SlaveFailed				  = 3
)

type ElevatorError struct {
	Code ErrorCode
}
