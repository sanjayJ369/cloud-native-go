package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKVStore(t *testing.T) {

	kvstore := NewKVStore()

	t.Run("test put", func(t *testing.T) {
		key, val := "hello", "world"
		err := kvstore.Put(key, val)
		assert.NoError(t, err)
	})

	t.Run("test get", func(t *testing.T) {

		testcases := []struct {
			name    string
			key     string
			value   string
			wantErr bool
		}{
			{
				name:    "existing value",
				key:     "hello",
				value:   "world",
				wantErr: false,
			}, {
				name:    "non exisint value",
				key:     "world",
				value:   "",
				wantErr: true,
			},
		}

		for _, tc := range testcases {
			value, err := kvstore.Get(tc.key)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.value, value)
			}
		}
	})

	t.Run("test del", func(t *testing.T) {
		testcases := []struct {
			name    string
			key     string
			value   string
			wantErr bool
		}{
			{
				name:    "exsitng value",
				key:     "hello",
				value:   "world",
				wantErr: false,
			}, {
				name:    "non existing value",
				key:     "world",
				value:   "",
				wantErr: true,
			},
		}

		for _, tc := range testcases {
			_, err := kvstore.Del(tc.key)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		}
	})
}
