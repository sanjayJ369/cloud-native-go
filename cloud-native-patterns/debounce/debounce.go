package main

import (
	"context"
	"sync"
	"time"
)

type Circuit func(context.Context) (string, error)

func externalService(ctx context.Context) (string, error) {
	return "", nil
}

func deboucneFirst(circut Circuit, d time.Duration) Circuit {
	var threshold time.Time
	var result string
	var err error
	var m sync.Mutex

	return func(ctx context.Context) (string, error) {
		m.Lock()
		defer m.Unlock()

		if time.Now().Before(threshold) {
			return result, err
		}

		result, err = externalService(ctx)

		threshold = time.Now().Add(d)
		return result, err
	}
}

func debounceFirstContext(circut Circuit, d time.Duration) Circuit {
	var threshold time.Time
	var cancel context.CancelFunc
	var result string
	var err error
	var m sync.Mutex

	return func(ctx context.Context) (string, error) {
		m.Lock()

		// cancel prev request
		if time.Now().Before(threshold) {
			cancel() // cancelling prev request
		}

		ctx, cancel = context.WithCancel(ctx)
		threshold = time.Now().Add(d)

		m.Unlock()

		// send a new request
		result, err = circut(ctx)

		return result, err
	}
}

func DebounceLate(circuit Circuit, d time.Duration) Circuit {

	var m sync.Mutex
	var timer *time.Timer
	var cancel context.CancelFunc
	var threshold time.Time

	return func(ctx context.Context) (string, error) {
		m.Lock()

		// cancel the timer
		// cancel the prev function request
		if timer != nil {
			timer.Stop()
			cancel()
		}

		cctx, cancel := context.WithCancel(ctx)
		ch := make(chan struct {
			result string
			err    error
		}, 1)

		timer = time.AfterFunc(d, func() {
			res, err := circuit(cctx)
			ch <- struct {
				result string
				err    error
			}{res, err}
		})

		m.Unlock()

		select {
		case res := <-ch:
			return res.result, res.err
		case <-cctx.Done():
			return "", cctx.Err()
		}
	}
}
