package main

import (
	"context"
	"log"
	"net"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/szymonwieloch/go-teleport/server/proto/teleportproto"
	"google.golang.org/grpc"
)

type server struct {
	teleportproto.UnimplementedRemoteExecutorServer
}

func (s *server) Start(ctx context.Context, req *teleportproto.Command) (*teleportproto.StartedTask, error) {
	log.Println("Starting command", req.Command)
	return &teleportproto.StartedTask{Id: &teleportproto.TaskId{Uuid: "123"}}, nil
}

func (s *server) Stop(ctx context.Context, req *teleportproto.TaskId) (*teleportproto.StoppedTask, error) {
	log.Println("Stopping job", req.Uuid)
	return &teleportproto.StoppedTask{}, nil
}

func (s *server) List(ctx context.Context, req *empty.Empty) (*teleportproto.ListOfTasks, error) {
	log.Println("Listing jobs")
	return &teleportproto.ListOfTasks{}, nil
}

func (s *server) Logs(req *teleportproto.TaskId, srv grpc.ServerStreamingServer[teleportproto.Log]) error {
	log.Println("Showing logs for job", req.Uuid)
	return nil
}

func (s *server) GetStatus(ctx context.Context, req *teleportproto.TaskId) (*teleportproto.Status, error) {
	log.Println("Showing status for job", req.Uuid)
	return &teleportproto.Status{}, nil
}

func startServer(addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	teleportproto.RegisterRemoteExecutorServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
