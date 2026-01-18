package main

import (
	"fmt"
	"go-micro/internal/api"
	db "go-micro/internal/store"
	tl "go-micro/internal/transationLogger"
	pb "go-micro/proto/store"
	"log"
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	s      db.Store
	logger tl.TransactionLogger
}

func NewServer(s db.Store, logger tl.TransactionLogger) *Server {
	err := tl.InitalizeTrasactionLogger(logger, s)
	if err != nil {
		log.Fatalf("error initalizting logger: %s", err)
	}

	return &Server{
		s:      s,
		logger: logger,
	}
}

func (s *Server) ListenAndServe(port int) error {
	const addr = "0.0.0.0"

	listener, err := net.Listen("tcp", net.JoinHostPort(addr, strconv.Itoa(port)))
	if err != nil {
		return fmt.Errorf("starting the server: %s", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterStoreServiceServer(grpcServer, &api.StoreServer{KVStore: s.s, Logger: s.logger})
	reflection.Register(grpcServer)
	err = grpcServer.Serve(listener)
	if err != nil {
		return fmt.Errorf("error servering server: %s", err)
	}

	return nil
}
