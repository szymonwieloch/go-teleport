package main

//go:generate mkdir -p proto mocks
//go:generate protoc -I=../../proto --go-grpc_out=. --go_out=. teleport.proto
//go:generate mockgen -destination mocks/grpc_mock.go -package mocks ./proto/teleportproto RemoteExecutorClient
//go:generate mockgen -destination mocks/grpc_steam_mock.go -package mocks google.golang.org/grpc ServerStreamingClient

import "fmt"

func main() {
	fmt.Println("Teleport client")
	args := parseArgs()
	execute(args)
}
