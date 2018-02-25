package main

import (
	"fmt"
	"elevatorAPI"
)


func main() {

	elevatorAPI.Init()

	fmt.Println("App started")
	blocker := make(chan bool, 1)
	<- blocker
}
