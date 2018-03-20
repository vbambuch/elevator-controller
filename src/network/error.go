package network

import (
	"consts"
	"time"
)

func ErrorDetection(errorChan chan<- consts.ElevatorError) {
	for {
		time.Sleep(50 * time.Second)
		errorChan <- consts.ElevatorError{consts.MasterFailed}

	}
	return
}
