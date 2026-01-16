package store

import "sync"

type KVStore struct {
	sync.RWMutex
	m map[string]string
}

func NewKVStore() *KVStore {
	return &KVStore{
		m: make(map[string]string),
	}
}

func (k *KVStore) Put(key, value string) error {
	k.Lock()
	defer k.Unlock()
	k.m[key] = value
	return nil
}

// returns val of the key
func (k *KVStore) Get(key string) (string, error) {
	k.RLock()
	defer k.RUnlock()

	val, ok := k.m[key]
	if !ok {
		return "", ErrorNoSuchKey
	}

	return val, nil
}

// return val that is being deleted and error
func (k *KVStore) Del(key string) (string, error) {
	k.Lock()
	defer k.Unlock()

	val, ok := k.m[key]
	if !ok {
		return "", ErrorNoSuchKey
	}

	delete(k.m, key)
	return val, nil
}
