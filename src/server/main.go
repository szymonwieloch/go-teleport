package main

//go:generate mkdir -p proto
//go:generate protoc -I=../../proto --go-grpc_out=. --go_out=. teleport.proto

import (
	"fmt"
	"log"

	"github.com/szymonwieloch/go-teleport/server/service"
)

func main() {
	fmt.Println("Teleport server")
	args := parseArgs()
	opts := service.ServiceOptions{
		Address:  args.Address,
		AuthKey:  args.AuthKey,
		AuthCert: args.AuthCert,
		Secret:   args.Secret,
		Limits:   args.Limits,
	}
	srv, err := service.NewService(opts)
	if err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
	defer srv.Close()
	err = srv.Serve()
	if err != nil {
		log.Fatalf("Error while serving: %v", err)
	}
}
