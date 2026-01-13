package main

import (
	"fmt"
	"sync"
	"time"
)

// chord
// chord takes in bunch of input channels
// produces an output channel
// here the value is put into the output channel
// only if a input has been received from every channel
// **note** the thing to be noted here is
// if two values comes from the same channel
// it ignores the inital value by updating it with
// the sencond value that has come from the channel

func Chord(sources ...<-chan int) <-chan []int {
	// create an intermediate type
	type Input struct {
		idx, value int
	}

	dest := make(chan []int)
	inputs := make(chan Input)
	var wg sync.WaitGroup // as usual to close the channel
	wg.Add(len(sources))

	// read data from source and put into
	// intermediate channel
	for i, ch := range sources {
		go func(idx int, ch <-chan int) {
			defer wg.Done()
			for val := range ch {
				inputs <- Input{idx: i, value: val}
			}
		}(i, ch)
	}

	go func() {
		wg.Wait()
		close(inputs)
	}()

	go func() {
		res := make([]int, len(sources))
		sent := make([]bool, len(sources))

		count := len(sources)
		for input := range inputs {
			res[input.idx] = input.value

			if !sent[input.idx] {
				sent[input.idx] = true
				count--
			}

			if count == 0 {
				// create copy of res
				// and send it to dest
				c := make([]int, len(sources))
				copy(c, res)
				dest <- c

				count = len(sources)
				clear(sent)
			}
		}
		close(dest)
	}()

	return dest
}

func main() {
	ch1 := make(chan int)
	ch2 := make(chan int)

	go func() {
		i := 0
		for {
			time.Sleep(time.Second)
			ch1 <- i
			i++
		}
	}()

	go func() {
		i := 0
		for {
			time.Sleep(time.Second * 2)
			ch2 <- i
			i++
		}
	}()

	res := Chord(ch1, ch2)

	for val := range res {
		fmt.Println(val)
	}
}
