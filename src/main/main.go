package main

import (
	"fmt"
	"elevator"
)


func main() {

	elevator.Init()

	fmt.Println("App started")
	blocker := make(chan bool, 1)
	<- blocker
}
