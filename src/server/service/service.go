package service

import (
	"fmt"
	"net"

	"github.com/szymonwieloch/go-teleport/server/proto/teleportproto"
	"google.golang.org/grpc"
)

type ServiceOptions struct {
	Address  string
	AuthKey  string
	AuthCert string
	Secret   string
	Limits   bool
}

type Service struct {
	server     *server
	grpcServer *grpc.Server
	listener   net.Listener
}

func (srv Service) Close() {
	srv.server.Close()
	srv.listener.Close()
}

func (srv Service) Serve() error {
	return srv.grpcServer.Serve(srv.listener)
}

func (srv Service) ServeInBackground() func() error {
	ch := make(chan error)
	go func() {
		ch <- srv.Serve()
	}()
	return func() error {
		srv.grpcServer.Stop()
		return <-ch
	}
}

// Starts server on the provided domain:port address
func NewService(args ServiceOptions) (Service, error) {
	opts := []grpc.ServerOption{}
	var err error
	if args.AuthKey != "" {
		opts, err = configOAuth(opts, args)
		if err != nil {
			return Service{}, fmt.Errorf("failed to configure authentication: %w", err)
		}
	}
	grpcServer := grpc.NewServer(opts...)
	lis, err := net.Listen("tcp", args.Address)
	if err != nil {
		return Service{}, fmt.Errorf("failed to listen: %w", err)
	}
	fmt.Println("Listening at", args.Address)
	defer func() {
		if lis != nil {
			lis.Close()
		}
	}()

	server, err := newServer(args)
	if err != nil {
		return Service{}, err
	}

	teleportproto.RegisterRemoteExecutorServer(grpcServer, server)
	service := Service{
		server:     server,
		grpcServer: grpcServer,
		listener:   lis,
	}
	// cancel defered close
	lis = nil
	return service, nil
}
