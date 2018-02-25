package consts


const (
	// writing
	ReloadConfig 		= "\x00"
	MotorDirection		= "\x01"
	OrderButtonLight	= "\x02"
	FloorIndicator		= "\x03"
	DoorOpenLight		= "\x04"
	StopButtonLight		= "\x05"

	// reading
	OrderButtonPressed	= "\x06"
	FloorSensor			= "\x07"
	StopButtonPressed	= "\x08"
	ObstructionSwitch	= "\x09"

	// misc
	EmptyByte			= "\x00"
	TurnOn				= "\x01"
	TurnOff				= "\x00"
	HallButtonUp		= "\x00"
	HallButtonDown		= "\x01"
	CabButton			= "\x02"
	MotorUp				= "\x0a"
	MotorDown			= "\xfe"
	MotorStop			= "\x00"
)


func getType(Type string) (string) {
	return Type
}
