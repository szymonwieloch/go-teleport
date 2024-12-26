//go:generate mkdir -p proto
//go:generate protoc -I=../../proto --go-grpc_out=. --go_out=. teleport.proto

// --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative

package main

import "fmt"

func main() {
	fmt.Println("Teleport client")
	args := parseArgs()
	execute(args)
}
