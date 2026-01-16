package transationlogger

const (
	EventPut int = iota
	EventDelete
)

type Event struct {
	Id        uint64 // event id: monotonically incereasing
	EventType int    // event type: put, delete
	Key       string
	Value     string
}

type TransactionLogger interface {
	WritePut(string, string)
	WriteDel(string)

	Err() <-chan error
	Run()
	ReadEvents() (<-chan Event, <-chan error) // stream the logged event in file
}
