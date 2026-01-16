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
	Del(string) (string, error)
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

	if errors.Is(err, ErrorNoSuchKey) {
		return res, status.Errorf(codes.NotFound, "key:%s not found", key)
	}
	if err != nil {
		return res, status.Errorf(codes.Internal, "internal server error: %s", err)
	}

	res.Value = val
	return res, nil
}

func (s *StoreServer) PutHandler(ctx context.Context, req *store.PutRequest) (*store.PutResponse, error) {
	key := req.GetKey()
	val := req.GetValue()
	res := &store.PutResponse{}
	err := s.KVStore.Put(key, val)
	if err != nil {
		return res, status.Errorf(codes.Internal, "internal server error: %s", err)
	}
	res.Key = key
	res.Value = val
	return res, nil
}

func (s *StoreServer) DelHandler(ctx context.Context, req *store.DelRequest) (*store.DelResponse, error) {
	key := req.GetKey()
	res := &store.DelResponse{}
	val, err := s.KVStore.Del(key)

	if errors.Is(err, ErrorNoSuchKey) {
		return res, status.Errorf(codes.NotFound, "key:%s not found", key)
	}
	if err != nil {
		return res, status.Errorf(codes.Internal, "internal server error: %s", err)
	}

	res.Key = key
	res.Value = val
	return res, nil
}
