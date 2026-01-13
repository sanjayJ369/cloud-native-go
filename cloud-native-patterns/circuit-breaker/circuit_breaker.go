package main

import "fmt"

// Patter circute breaker
// you create an adapter(breaker), that wraps around a external dependency(circut)


// User
type User struct {
	Id int
	// some other user info
}

// circut
type Circut func(context.Context) (string, error)

// breaker
// takes in the circut and a threshold
// keep track of the number of failus
// if the number of failures is like greater
// then a certain threshold
// trip the breaker
func breaker(circuit Circut, threshold int) Circut {
	var failures int
	var last = time.Now()
	var m sync.RWMutex
	return func(ctx context.Contex) (string, error) {
		m.RLock()

		if failures > threshold {
			// the external dependency is broken
			// return an error
			
			retry := last.Add(2 * time.Minute)
			// wait 2min until next try 
			if !time.Now().After(retry) {
				m.RUnlock()
				return "", errors.New("service unavilable")
			}
		}

		m.RUnlock()

		resp, err := circut(ctx)

		m.Lock()
		defer m.Unlock()

		last = time.Now()
		
		if err != nil {
			failures++
			return resp, err
		}

		failures = 0
		return resp, nil
	}
}
		

		
			
			
	

	



	

