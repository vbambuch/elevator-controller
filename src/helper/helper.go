package helper

import "time"

func GetInstruction(b0, b1, b2, b3 byte) ([]byte){
	return []byte{b0, b1, b2, b3}
}


func Timeout(ms time.Duration, timeout chan<- bool) {
	time.Sleep(ms * time.Millisecond)
	timeout <- true
}
