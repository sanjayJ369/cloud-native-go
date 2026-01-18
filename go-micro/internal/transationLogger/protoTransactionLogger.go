package transationlogger

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	protobufLogger "go-micro/proto/transactionLogger"
	"io"
	"log"
	"os"
	"sync/atomic"

	"google.golang.org/protobuf/proto"
)

type ProtoTransactionLogger struct {
	events      chan<- Event
	errors      <-chan error
	lastEventId uint64
	file        *os.File
}

func NewProtoTransactionLogger(filename string) (TransactionLogger, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("error creating file %s: %s", filename, err)
	}

	return &ProtoTransactionLogger{
		file: file,
	}, nil
}

func (p *ProtoTransactionLogger) WritePut(key, value string) {
	p.events <- Event{EventType: EventPut, Key: key, Value: value}
}

func (p *ProtoTransactionLogger) WriteDel(key string) {
	p.events <- Event{EventType: EventDelete, Key: key}
}

func (p *ProtoTransactionLogger) Err() <-chan error {
	return p.errors
}

func (p *ProtoTransactionLogger) Run() {
	eventChan := make(chan Event, 16)
	p.events = eventChan

	errorChan := make(chan error, 1)
	p.errors = errorChan

	go func() {
		writer := bufio.NewWriter(p.file)
		datalen := make([]byte, 4)
		event := &protobufLogger.Event{}
		for e := range eventChan {
			atomic.AddUint64(&p.lastEventId, 1)

			event.Id = atomic.LoadUint64(&p.lastEventId)
			event.EventType = uint32(e.EventType)
			event.Key = e.Key
			event.Value = e.Value

			data, err := proto.Marshal(event)
			if err != nil {
				errorChan <- fmt.Errorf("error marshaling event: %s", err)
				return
			}

			binary.LittleEndian.PutUint32(datalen, uint32(len(data)))
			_, err = writer.Write(datalen)
			if err != nil {
				errorChan <- fmt.Errorf("error writing data len: %s", err)
				return
			}

			_, err = writer.Write(data)
			if err != nil {
				errorChan <- fmt.Errorf("error writing data: %s", err)
				return
			}

			err = writer.Flush()
			if err != nil {
				errorChan <- fmt.Errorf("error flushing data: %s", err)
				return
			}
		}
	}()
}

func (p *ProtoTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	outEvent := make(chan Event)
	outError := make(chan error, 1)

	go func() {
		defer close(outEvent)
		defer close(outError)

		atomic.StoreUint64(&p.lastEventId, 0)

		file, err := os.OpenFile(p.file.Name(), os.O_RDWR, 0755)
		if err != nil {
			outError <- fmt.Errorf("error creating file %s: %s", p.file.Name(), err)
			return
		}
		reader := bufio.NewReader(file)

		lenbuf := make([]byte, 4)
		databuf := make([]byte, 4096)
		for {

			_, err := io.ReadFull(reader, lenbuf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				outError <- fmt.Errorf("error reading event entry length: %s", err)
				return
			}

			datalen := binary.LittleEndian.Uint32(lenbuf)
			if datalen > 4096 {
				// todo: add support for bigger event logs
				log.Fatalln("plz implement support for larger buffers ;)")
			}

			_, err = io.ReadFull(reader, databuf[:datalen])
			if err != nil {
				outError <- fmt.Errorf("error reading event entry: %s", err)
				return
			}

			event := &protobufLogger.Event{}
			err = proto.Unmarshal(databuf[:datalen], event)
			if err != nil {
				outError <- fmt.Errorf("error unmarshalling event entry: %s", err)
				return
			}

			if atomic.LoadUint64(&p.lastEventId) >= event.Id {
				outError <- fmt.Errorf("invalid sequence number")
				return
			}

			atomic.StoreUint64(&p.lastEventId, event.Id)

			outEvent <- Event{
				Id:        event.Id,
				EventType: int(event.EventType),
				Key:       event.Key,
				Value:     event.Value,
			}
		}
	}()

	return outEvent, outError
}

func (p *ProtoTransactionLogger) GetLastEventId() uint64 {
	return atomic.LoadUint64(&p.lastEventId)
}
