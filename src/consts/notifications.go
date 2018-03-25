package consts

import "encoding/json"

// Notifications
type Notification []byte
type notificationCode int

const (
	SlavePeriodicMsg  notificationCode = 0
	SlaveHallOrder                     = 1
	SlaveReady						   = 5

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
	Floor     int
	Direction MotorDirection
	CabArray  []ButtonEvent
	Ready     bool
}
