package main

import (
	"context"
	"errors"
	"go/token"
	"sync"
	"time"
)

type Effector func(context.Context) (string, error)

func ThrottleSimple(effector Effector, count int, d time.Duration) Effector {
	var nextReset time.Time
	var countLeft int = count
	var m sync.Mutex

	return func(ctx context.Context) (string, error) {
		m.Lock()
		defer m.Unlock()

		if time.Now().After(nextReset) {
			countLeft = count
			nextReset = time.Now().Add(d)
		}

		if count > 0 {
			count--
			result, err := effector(ctx)
			return result, err
		}

		return "", errors.New("request limit reached")
	}
}

func Throttle(e Effector, maxTokens, refill int, d time.Duration) Effector {
	var tokens int = maxTokens
	var m sync.Mutex
	var once sync.Once

	return func(ctx context.Context) (string, error) {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		once.Do(func() {
			ticker := time.NewTicker(d)

			go func() {
				defer ticker.Stop()

				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						m.Lock()
						tokens += refill
						if tokens > maxTokens {
							tokens = maxTokens
						}
					}
				}
			}()
		})

		m.Lock()
		defer m.Unlock()

		if tokens <= 0 {
			return "", errors.New("too many requests")
		}

		tokens--

		return e(ctx)
	}
}
