package api

import (
	"context"
	"errors"
	"go-micro/internal/store"
	tl "go-micro/internal/transationLogger"
	pb "go-micro/proto/store"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StoreServer struct {
	pb.UnimplementedStoreServiceServer
	KVStore store.Store
	Logger  tl.TransactionLogger
}

func (s *StoreServer) GetHandler(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	key := req.GetKey()
	res := &pb.GetResponse{Value: ""}
	val, err := s.KVStore.Get(key)

	if errors.Is(err, store.ErrorNoSuchKey) {
		return res, status.Errorf(codes.NotFound, "key:%s not found", key)
	}
	if err != nil {
		return res, status.Errorf(codes.Internal, "internal server error: %s", err)
	}

	res.Value = val
	return res, nil
}

func (s *StoreServer) PutHandler(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	key := req.GetKey()
	val := req.GetValue()
	res := &pb.PutResponse{}

	// write to inmem store
	err := s.KVStore.Put(key, val)
	if err != nil {
		return res, status.Errorf(codes.Internal, "internal server error: %s", err)
	}

	// write to logger
	s.Logger.WritePut(key, val)

	res.Key = key
	res.Value = val
	return res, nil
}

func (s *StoreServer) DelHandler(ctx context.Context, req *pb.DelRequest) (*pb.DelResponse, error) {
	key := req.GetKey()
	res := &pb.DelResponse{}
	val, err := s.KVStore.Del(key)

	if errors.Is(err, store.ErrorNoSuchKey) {
		return res, status.Errorf(codes.NotFound, "key:%s not found", key)
	}
	if err != nil {
		return res, status.Errorf(codes.Internal, "internal server error: %s", err)
	}

	// write to db
	s.Logger.WriteDel(key)

	res.Key = key
	res.Value = val
	return res, nil
}
