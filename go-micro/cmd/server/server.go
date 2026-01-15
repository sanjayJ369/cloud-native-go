package main

import (
	"fmt"
	db "go-micro/internal/store"
	pb "go-micro/proto/store"
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	s db.Store
}

func NewServer(s db.Store) *Server {
	return &Server{
		s: s,
	}
}
func (s *Server) ListenAndServe(port int) error {
	const addr = "0.0.0.0"

	listener, err := net.Listen("tcp", net.JoinHostPort(addr, strconv.Itoa(port)))
	if err != nil {
		return fmt.Errorf("starting the server: %s", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterStoreServiceServer(grpcServer, &db.StoreServer{KVStore: s.s})
	reflection.Register(grpcServer)
	err = grpcServer.Serve(listener)
	if err != nil {
		return fmt.Errorf("error servering server: %s", err)
	}

	return nil
}
