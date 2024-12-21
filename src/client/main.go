//go:generate mkdir -p proto
//go:generate protoc -I=../../proto --go_out=. teleport.proto

package main

import "fmt"

func main() {
	fmt.Println("Teleport client")
}
