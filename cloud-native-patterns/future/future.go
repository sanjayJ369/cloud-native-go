package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Future interface {
	Result() (string, error)
}

type InnerFuture struct {
	wg   sync.WaitGroup
	once sync.Once

	res string
	err error

	resch <-chan string
	errch <-chan error
}

func (i *InnerFuture) Result() (string, error) {
	i.once.Do(func() {
		i.wg.Add(1)
		defer i.wg.Done()
		i.res = <-i.resch
		i.err = <-i.errch
	})

	i.wg.Wait()
	return i.res, i.err
}

func SlowFunction(ctx context.Context) Future {
	resChan := make(chan string)
	errChan := make(chan error)

	go func() {
		select {
		case <-time.After(2 * time.Second):
			resChan <- "doing slow work"
			errChan <- nil
		case <-ctx.Done():
			resChan <- ""
			errChan <- ctx.Err()
		}
	}()

	return &InnerFuture{resch: resChan, errch: errChan}
}

func main() {

	start := time.Now()
	res1 := SlowFunction(context.Background())
	res2 := SlowFunction(context.Background())

	fmt.Println(res1.Result())
	fmt.Println(res2.Result())

	fmt.Println(time.Since(start))

}
