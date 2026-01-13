package main

// channel type
type Data int

// fanout takes in a source and returns
// mulitple destination channels
// where the input from the source will be split into
// all the dsetination channels uniformly
func fanout(source <-chan Data, count int) []<-chan Data {
	var dest []<-chan Data

	for range count {
		ch := make(chan Data)
		dest = append(dest, ch)

		go func() {
			defer close(ch)
			for val := range source {
				ch <- val
			}
		}()
	}

	return dest
}
