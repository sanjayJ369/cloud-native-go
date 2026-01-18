package main

import (
	"fmt"
	db "go-micro/internal/store"
	tl "go-micro/internal/transationLogger"
	"log"
)

func main() {
	store := db.NewKVStore()
	logger, err := tl.NewProtoTransactionLogger("./transaction.log")
	if err != nil {
		log.Fatalln(err)
	}

	srv := NewServer(store, logger)
	err = srv.ListenAndServe(8080)
	if err != nil {
		fmt.Println("error while running the server: %s", err)
	}
}
