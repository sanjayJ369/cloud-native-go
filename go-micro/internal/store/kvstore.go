package store

type KVStore struct {
	m map[string]string
}

func NewKVStore() *KVStore {
	return &KVStore{
		m: make(map[string]string),
	}
}

func (k *KVStore) Put(key, value string) error {
	k.m[key] = value
	return nil
}

// returns val of the key
func (k *KVStore) Get(key string) (string, error) {
	val, ok := k.m[key]
	if !ok {
		return "", ErrorNoSuchKey
	}

	return val, nil
}

// return val that is being deleted and error
func (k *KVStore) Del(key string) (string, error) {
	val, ok := k.m[key]
	if !ok {
		return "", ErrorNoSuchKey
	}

	delete(k.m, key)
	return val, nil
}
