package transactionLogger

import "go-micro/internal/store"

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
	GetLastEventId() uint64                   // retuns the number of events written to the file
}

func InitalizeTrasactionLogger(logger TransactionLogger, store store.Store) error {
	var err error

	events, errors := logger.ReadEvents()
	e := Event{}
	ok := true

	// read events into in-mem store
	for ok && err == nil {
		select {
		case err, ok = <-errors:
		case e, ok = <-events:
			switch e.EventType {
			case EventDelete:
				store.Del(e.Key)
			case EventPut:
				store.Put(e.Key, e.Value)
			}
		}
	}

	logger.Run()
	return nil

}
