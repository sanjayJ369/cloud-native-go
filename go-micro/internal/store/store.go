package store

import (
	"context"
	"errors"
	"go-micro/proto/store"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// globals

// basic key value store interface
type Store interface {
	Put(string, string) error
	Get(string) (string, error)
	Del(string) error
}

var ErrorNoSuchKey = errors.New("no such key")

type StoreServer struct {
	store.UnimplementedStoreServiceServer
	KVStore Store
}

func (s *StoreServer) GetHandler(ctx context.Context, req *store.GetRequest) (*store.GetResponse, error) {
	key := req.GetKey()
	res := &store.GetResponse{Value: ""}
	val, err := s.KVStore.Get(key)
	if err != nil {
		return res, status.Errorf(codes.NotFound, "key:%s not found", key)
	}

	res.Value = val
	return res, nil
}
