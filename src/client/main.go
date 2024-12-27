package main

//go:generate mkdir -p proto
//go:generate protoc -I=../../proto --go-grpc_out=. --go_out=. teleport.proto

import "fmt"

func main() {
	fmt.Println("Teleport client")
	args := parseArgs()
	execute(args)
}
