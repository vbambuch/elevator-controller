package consts

const MinFloor = 0
const MaxFloor = 3

const Address = "localhost"
const Port = "15657"



type Elevator struct {
	Status int
	Floor int
	AtFloor bool
	OrderButton bool
	StopButton bool
	Obstruction bool
}
