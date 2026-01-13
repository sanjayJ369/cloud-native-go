package main

import (
	"fmt"
	"sync"
	"time"
)

// worker pool pattern has 3 main things
// workers - which do the job
// jobs channel - which includes the job to be done
// result channel - results from the worker pool

type Job int

func worker(id int, jobs <-chan Job, result chan<- Job) {
	for job := range jobs {
		// do work
		fmt.Printf("worker %d, start job", job)
		time.Sleep(time.Second)
		result <- job * 2
	}
}

func main() {
	jobs := make(chan Job, 10)
	res := make(chan Job)

	var wg sync.WaitGroup

	// spin up workers
	for i := range 3 {
		go worker(i, jobs, res)
	}

	// send 10 jobs
	for i := range 10 {
		wg.Add(1)
		jobs <- Job(i)
	}

	go func() {
		wg.Wait()
		close(jobs)
		close(res)
	}()

	for val := range res {
		wg.Done()
		fmt.Println(val)
	}
}
