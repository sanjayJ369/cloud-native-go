package main

import (
	"context"
	"time"
)

type Effector func(context.Context) (string, error)

func Retry(effector Effector, retries int, delay time.Duration) Effector {

	return func(ctx context.Context) (string, error) {
		for r := 1; ; r++ {

			res, err := effector(ctx)
			if err == nil || r >= retries {
				return res, err
			}

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
	}
}
