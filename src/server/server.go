package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/szymonwieloch/go-teleport/server/jobs"
	"github.com/szymonwieloch/go-teleport/server/proto/teleportproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	teleportproto.UnimplementedRemoteExecutorServer
	jobs *jobs.Jobs
}

func NewServer() *server {
	return &server{jobs: jobs.NewJobs()}
}

func (s *server) Start(ctx context.Context, req *teleportproto.Command) (*teleportproto.JobStatus, error) {
	log.Println("Starting command", req.Command)
	job, err := s.jobs.Create(req.Command)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not start the process")
	}
	return jobStatus(job.Status()), nil
}

func (s *server) Stop(ctx context.Context, req *teleportproto.JobId) (*teleportproto.JobStatus, error) {
	log.Println("Stopping job", req.Uuid)
	job, err := s.jobs.Stop(jobs.JobID(req.Uuid))
	if err != nil {
		if err == jobs.ErrNotFound {
			return nil, status.Error(codes.NotFound, "id was not found")
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return jobStatus(job.Status()), nil

}

func (s *server) List(ctx context.Context, req *empty.Empty) (*teleportproto.JobList, error) {
	log.Println("Listing jobs")
	jobs := s.jobs.List()
	output := make([]*teleportproto.JobStatus, 0, len(jobs))
	for _, job := range jobs {
		output = append(output, jobStatus(job.Status()))
	}
	return &teleportproto.JobList{Jobs: output}, nil
}

func (s *server) Logs(req *teleportproto.JobId, srv grpc.ServerStreamingServer[teleportproto.Log]) error {
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
			src := teleportproto.LogSource_LS_STDERR
			if log.Stdout {
				src = teleportproto.LogSource_LS_STDOUT
			}
			msg := &teleportproto.Log{
				Text:      log.Line,
				Src:       src,
				Timestamp: timestamppb.New(log.Timestamp),
			}
			err := srv.Send(msg)
			if err != nil {
				return err
			}
		}
	}
}

func (s *server) GetStatus(ctx context.Context, req *teleportproto.JobId) (*teleportproto.JobStatus, error) {
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
	fmt.Println("Server started at", addr)
	s := grpc.NewServer()
	teleportproto.RegisterRemoteExecutorServer(s, NewServer())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
