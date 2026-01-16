package transationlogger

import (
	"bufio"
	"fmt"
	"os"
)

type FileTransactionLogger struct {
	events      chan<- Event
	errors      <-chan error
	lastEventId uint64
	file        *os.File
}

func NewFileTransactionLogger(filename string) (TransactionLogger, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("error creating file %s: %s", filename, err)
	}
	return &FileTransactionLogger{file: file}, nil
}

func (f *FileTransactionLogger) WritePut(key, value string) {
	f.events <- Event{EventType: EventPut, Key: key, Value: value}
}

func (f *FileTransactionLogger) WriteDel(key string) {
	f.events <- Event{EventType: EventDelete, Key: key}
}

func (f *FileTransactionLogger) Err() <-chan error {
	return f.errors
}

// Run function spings up go routine
// to read the data from the events channel and write to file
func (f *FileTransactionLogger) Run() {
	events := make(chan Event, 16)
	f.events = events

	errors := make(chan error, 1)
	f.errors = errors

	go func() {
		for event := range events {
			f.lastEventId++
			_, err := fmt.Fprintf(f.file, "%d\t%s\t%s\t%s\n", f.lastEventId, event.EventType, event.Key, event.Value)
			if err != nil {
				errors <- err
				return
			}
		}
	}()
}

func (f *FileTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	scanner := bufio.NewScanner(f.file)

	outEvent := make(chan Event)
	outError := make(chan error, 1)

	go func() {
		var e Event

		defer close(outEvent)
		defer close(outError)

		for scanner.Scan() {
			line := scanner.Text()

			if _, err := fmt.Sscan(line, "%d\t%s\t%s\t%s\n", &e.Id, &e.EventType, &e.Key, &e.Value); err != nil {
				outError <- err
				return
			}

			if f.lastEventId >= e.Id {
				outError <- fmt.Errorf("invalid sequence number")
				return
			}

			f.lastEventId = e.Id

			outEvent <- e
		}
	}()

	return outEvent, outError
}
