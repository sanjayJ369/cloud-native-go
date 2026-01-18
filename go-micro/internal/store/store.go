package store

import (
	"errors"
)

// globals

// basic key value store interface
type Store interface {
	Put(string, string) error
	Get(string) (string, error)
	Del(string) (string, error)
}

var ErrorNoSuchKey = errors.New("no such key")
