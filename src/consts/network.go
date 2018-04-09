package consts


//Message struct
type Message struct {
	Category int
	Master   bool
	Floor    int
	Button   int
	Addr     string //`json:"-"`
}

//Local ListenIP address
var Laddr string

//Channel for closing connections
var CloseConnChan = make(chan bool)

// Message category constant
const (
	Alive int = iota + 1

)
