package main

//go:generate mkdir -p proto
//go:generate protoc -I=../../proto --go-grpc_out=. --go_out=. teleport.proto

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("Teleport server")
	args := parseArgs()
	err := startServer(args)
	if err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
