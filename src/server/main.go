//go:generate mkdir -p proto
//go:generate protoc -I=../../proto --go-grpc_out=. --go_out=. teleport.proto

package main

import "fmt"

func main() {
	fmt.Println("Teleport server")
	args := parseArgs()
	startServer(args.Address)

}
