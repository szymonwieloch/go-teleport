package main

import (
	"context"
	"log"
	"net"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/szymonwieloch/go-teleport/server/jobs"
	"github.com/szymonwieloch/go-teleport/server/proto/teleportproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	teleportproto.UnimplementedRemoteExecutorServer
	jobs jobs.Jobs
}

func (s *server) Start(ctx context.Context, req *teleportproto.Command) (*teleportproto.StartedTask, error) {
	log.Println("Starting command", req.Command)
	job, err := s.jobs.Create(req.Command)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not start the process")
	}
	running, _ := job.Status()
	if running == nil {
		return nil, status.Error(codes.Internal, "process not running")
	}
	return startedJobStatus(running), nil
}

func (s *server) Stop(ctx context.Context, req *teleportproto.TaskId) (*teleportproto.StoppedTask, error) {
	log.Println("Stopping job", req.Uuid)
	job, err := s.jobs.Stop(jobs.JobID(req.Uuid))
	if err != nil {
		if err == jobs.ErrNotFound {
			return nil, status.Error(codes.NotFound, "id was not found")
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	_, stopped := job.Status()
	if stopped == nil {
		return nil, status.Error(codes.Internal, "process not stopped")
	}
	return stoppedJobStatus(stopped), nil
}

func (s *server) List(ctx context.Context, req *empty.Empty) (*teleportproto.ListOfTasks, error) {
	log.Println("Listing jobs")
	jobs := s.jobs.List()
	output := make([]*teleportproto.Status, 0, len(jobs))
	for _, job := range jobs {
		output = append(output, jobStatus(job.Status()))
	}
	return &teleportproto.ListOfTasks{Tasks: output}, nil
}

func (s *server) Logs(req *teleportproto.TaskId, srv grpc.ServerStreamingServer[teleportproto.Log]) error {
	log.Println("Showing logs for job", req.Uuid)
	job := s.jobs.Find(jobs.JobID(req.Uuid))
	if job == nil {
		return status.Error(codes.NotFound, "id was not found")
	}
	position := 0
	for {
		logs := job.GetLogs(position, 10)
		if len(logs) == 0 {
			return nil
		}
		position += len(logs)
		for _, log := range logs {
			if err := srv.Send(&teleportproto.Log{Stdout: &teleportproto.TextOutput{Text: log}}); err != nil {
				return err
			}
		}
	}
}

func (s *server) GetStatus(ctx context.Context, req *teleportproto.TaskId) (*teleportproto.Status, error) {
	log.Println("Showing status for job", req.Uuid)
	job := s.jobs.Find(jobs.JobID(req.Uuid))
	if job == nil {
		return nil, status.Error(codes.NotFound, "id was not found")
	}
	return jobStatus(job.Status()), nil
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
