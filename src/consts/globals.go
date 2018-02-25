package consts

const MinFloor = 0
const MaxFloor = 3

const Address = "localhost"
const Port = "15657"



type elevator struct {
	Status int
	Floor int
	AtFloor bool
	OrderButton bool
	StopButton bool
	Obstruction bool
}


type Elevator struct {
	Data elevator
	Test elevator
}
