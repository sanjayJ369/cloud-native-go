package main

import (
	"fmt"
	db "go-micro/internal/store"
)

func main() {
	store := db.NewKVStore()

	srv := NewServer(store)
	err := srv.ListenAndServe(8080)
	if err != nil {
		fmt.Println("error while running the server: %s", err)
	}
}
