package transactionLogger

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
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

func (f *FileTransactionLogger) GetLastEventId() uint64 {
	return atomic.LoadUint64(&f.lastEventId)
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
			atomic.AddUint64(&f.lastEventId, 1)
			_, err := fmt.Fprintf(f.file, "%d\t%d\t%s\t%s\n", f.lastEventId, event.EventType, event.Key, event.Value)
			if err != nil {
				errors <- err
				return
			}
		}
	}()
}

func (f *FileTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	outEvent := make(chan Event)
	outError := make(chan error, 1)

	go func() {
		e := Event{}

		file, err := os.OpenFile(f.file.Name(), os.O_RDWR, 0755)
		if err != nil {
			outError <- fmt.Errorf("error creating file %s: %s", f.file.Name(), err)
			return
		}

		scanner := bufio.NewScanner(file)

		defer close(outEvent)
		defer close(outError)

		atomic.StoreUint64(&f.lastEventId, 0)

		for scanner.Scan() {
			line := scanner.Text()

			contents := strings.Split(line, "\t")
			if num, err := strconv.Atoi(contents[1]); num == EventPut {
				if err != nil {
					outError <- fmt.Errorf("invalid event type: %s", err)
					return
				}

				if _, err := fmt.Sscanf(line, "%d\t%d\t%s\t%s", &e.Id, &e.EventType, &e.Key, &e.Value); err != nil {
					outError <- err
					return
				}
			} else {
				if _, err := fmt.Sscanf(line, "%d\t%d\t%s\t", &e.Id, &e.EventType, &e.Key); err != nil {
					outError <- err
					return
				}
			}

			if atomic.LoadUint64(&f.lastEventId) >= e.Id {
				outError <- fmt.Errorf("invalid sequence number")
				return
			}

			atomic.StoreUint64(&f.lastEventId, e.Id)

			outEvent <- e
		}

		if err := scanner.Err(); err != nil {
			outError <- fmt.Errorf("error reading file: %s", err)
			return
		}

	}()

	return outEvent, outError
}
