package main

import (
	"context"
	"time"
)

// function that takes a lot of time to run
type SlowFunction func(string) (string, error)

type WithContext func(context.Context, string) (string, error)

func timeout(f SlowFunction, ctx context.Context) WithContext {
	return func(ctx context.Context, s string) (string, error) {
		ch := make(chan struct {
			result string
			err    error
		}, 1)

		go func() {
			res, err := f(s)
			ch <- struct {
				result string
				err    error
			}{res, err}
		}()

		select {
		case res := <-ch:
			return res.result, res.err
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}

func WithTimeout(f SlowFunction, d time.Duration) WithContext {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()

	return timeout(f, ctx)
}
