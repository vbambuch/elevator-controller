package consts

import (
	"encoding/json"
)

// Notifications
type Notification []byte
type notificationCode int

const (
	SlavePeriodicMsg  notificationCode = 0
	SlaveHallOrder                     = 1
	ClearHallOrder					   = 5

	MasterHallLight                    = 2
	MasterHallOrder                    = 3
	MasterBroadcastIP                  = 4
)

type NotificationData struct {
	Code      notificationCode
	Data 	  json.RawMessage	// to be able to send different structures inside Data
}

// one type of data sending inside "NotificationData.Data"
type PeriodicData struct {
	ListenIP       string
	Floor          int
	Direction      MotorDirection
	OrderArray     []ButtonEvent
	Free           bool // elevator stopped on a floor and has empty cab call array
	HallProcessing bool // elevator is processing hall order
}
