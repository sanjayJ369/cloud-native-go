package main

import "sync"

// struct of the channel
type Data int

// funnel takes in a bunch of channels as input
// outputs a channel where combines results
// from all the input channels
func Funnel(sources ...<-chan Data) <-chan Data {
	var wg sync.WaitGroup
	var dest chan Data = make(chan Data)
	wg.Add(len(sources))

	for _, ch := range sources {
		go func(ch <-chan Data) {
			defer wg.Done()
			for n := range ch {
				dest <- n
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(dest)
	}()

	return dest
}
