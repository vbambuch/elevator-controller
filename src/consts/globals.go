package consts

import (
	"time"
	"net"
)

const MinFloor = 0
var NumFloors int
var MaxFloor int
const MiddleFloor = -1

const LocalAddress = "localhost:"
const MasterPort = "56789"
const BackupPort = "57890"
var ElevatorPort = ""

// Elevator consts
const DefaultValue = -2
var DefaultOrder = ButtonEvent{DefaultValue, DefaultValue}
const PollRate = 20 * time.Millisecond
const Unassigned = "UnassignedHallOrder"
const NoOutdated  = "NoOutdatedElevator"

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
	Backup		 = 2
	Slave 		 = 3
	DefaultRole  = 4
)

type HallOrders struct {
	Order      ButtonEvent
	AssignedTo string
}

// SlaveDB
type FreeElevatorItem struct {
	FloorDiff float64
	Data      DBItem
}

type DBItem struct {
	ClientConn 	*net.UDPConn
	Ignore     	int //ignore number of incoming messages
	Timestamp	time.Time
	Data       	PeriodicData
}

// Backup sync data
type BackupSync struct {
	SlavesList []DBItem
	OrderList  []HallOrders
	Timestamp  time.Time
}

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

// Logprint colours for printing to console
const (
	Grey = "\x1b[30;1m" // Dark grey
	Red = "\x1b[31;1m" // Red
	Green = "\x1b[32;1m" // Green
	Yellow = "\x1b[33;1m" // Yellow
	Blue = "\x1b[34;1m" // Blue
	Magenta = "\x1b[35;1m" // Magenta
	Cyan = "\x1b[36;1m" // Cyan
	White = "\x1b[37;1m" // White
	Neutral = "\x1b[0m"    // Grey (neutral)
)




