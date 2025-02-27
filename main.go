package main

import (
	"fmt"
	"simpleKV/server"
	"simpleKV/server/store"
)

func main() {
	store := store.NewStore(4, 4096)

	srv := server.NewServer(":6379", store)

	fmt.Println("Starting Redis-like server...")

	err := srv.Run()
	if err != nil {
		fmt.Println("Server error:", err)
	}
}
