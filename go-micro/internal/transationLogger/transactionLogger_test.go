package transactionLogger

import (
	"fmt"
	"go-micro/utils"
	"math/rand/v2"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func BenchmarkTransactionLogger(b *testing.B) {
	count := 1000
	tests := []struct {
		name    string
		factory func(string) (TransactionLogger, error)
	}{
		{
			name:    "file trasaction logger",
			factory: NewFileTransactionLogger,
		}, {
			name:    "proto transaction logger",
			factory: NewProtoTransactionLogger,
		},
	}

	// generate events
	events, _ := GenerateEvents(count)

	for _, tc := range tests {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				tempfile := filepath.Join(os.TempDir(), uuid.NewString()+".log")
				fl, _ := tc.factory(tempfile)
				fl.Run()
				b.StartTimer()

				// write events
				for _, e := range events {
					switch e.EventType {
					case EventDelete:
						fl.WriteDel(e.Key)
					case EventPut:
						fl.WritePut(e.Key, e.Value)
					default:
						b.Fatalf("wrong event type: %d", e.EventType)
					}
				}

				// wait
				for fl.GetLastEventId() < uint64(count) {
					runtime.Gosched()
				}

				// rebuild logs
				eventchan, _ := fl.ReadEvents()
				for _ = range eventchan {
				}

				b.StopTimer()
				os.Remove(tempfile)
				b.StartTimer()
			}

		})
	}
}

func TestTransactionLogger(t *testing.T) {

	count := 50
	tests := []struct {
		name    string
		factory func(string) (TransactionLogger, error)
	}{
		{
			name:    "string logger",
			factory: NewFileTransactionLogger,
		},
		{
			name:    "proto logger",
			factory: NewProtoTransactionLogger,
		},
	}

	for _, tc := range tests {
		events, finalMap := GenerateEvents(count)

		tempDir := os.TempDir()
		tempFile := fmt.Sprintf("%s/%s%s", tempDir, uuid.NewString(), ".txt")

		fl, err := tc.factory(tempFile)
		assert.NoError(t, err)

		fl.Run()
		for _, e := range events {
			switch e.EventType {
			case EventDelete:
				fl.WriteDel(e.Key)
			case EventPut:
				fl.WritePut(e.Key, e.Value)
			default:
				t.Fatalf("wrong event type: %d", e.EventType)
			}
		}
		// wait for the events to be written
		for fl.GetLastEventId() < uint64(count) {
			time.Sleep(time.Millisecond)
		}

		// rebuild map
		eventChan, errorChan := fl.ReadEvents()
		go func() {
			err := <-errorChan
			if err != nil {
				t.Errorf("error reading events: %s", err)
			}
		}()

		gotMap := make(map[string]string)
		for event := range eventChan {
			switch event.EventType {
			case EventDelete:
				delete(gotMap, event.Key)
			case EventPut:
				gotMap[event.Key] = event.Value
			default:
				t.Errorf("wrong event type: %d", event.EventType)
			}
		}

		// compare
		assert.Equal(t, finalMap, gotMap)
	}
}

// GenerateEvents generate random events and
// returns slice of event and a map represeting
// final state of the map
func GenerateEvents(count int) ([]Event, map[string]string) {
	// generate del event with 0.2 probability
	delProb := 0.4
	data := make(map[string]string)
	events := make([]Event, 0, count)

	var event Event
	for range count {
		delEvent := false

		if rand.Float32() < float32(delProb) && len(data) > 0 {
			delEvent = true
		}

		if delEvent {
			event = *GeneratePopDelEvent(data)
			events = append(events, event)
		} else {
			event = *GeneratePutEvent()
			events = append(events, event)
			data[event.Key] = event.Value
		}
	}

	return events, data
}

func GeneratePutEvent() *Event {
	maxkeylen := 50
	maxvaluelen := 500

	klen := rand.IntN(maxkeylen) + 1
	vlen := rand.IntN(maxvaluelen) + 1

	return &Event{
		EventType: EventPut,
		Key:       utils.RandomString(klen),
		Value:     utils.RandomString(vlen),
	}
}

func GeneratePopDelEvent(data map[string]string) *Event {
	for k, _ := range data {
		delete(data, k)
		return &Event{
			EventType: EventDelete,
			Key:       k,
		}
	}

	return nil
}
