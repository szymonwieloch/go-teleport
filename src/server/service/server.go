// Implementation of the gRPC server
package service

import (
	"context"
	"fmt"
	"log"

	"github.com/containerd/cgroups/v3/cgroup2"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/szymonwieloch/go-teleport/server/jobs"
	"github.com/szymonwieloch/go-teleport/server/proto/teleportproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// The main server type
type server struct {
	teleportproto.UnimplementedRemoteExecutorServer
	jobs *jobs.Jobs
}

// Creates a new instant of a server
func newServer(args ServiceOptions) (*server, error) {
	var cg *cgroup2.Manager
	var err error
	if args.Limits {
		cg, err = jobs.GetOrCreateGroup()
		if err != nil {
			return nil, fmt.Errorf("could not create cgroup: %w", err)
		}
	}
	j := jobs.NewJobs(cg)
	return &server{jobs: j}, nil
}

func (s *server) Close() {
	s.jobs.KillAll()
}

// The following is implementation of the teleportproto.RemoteExecutorServer interface
var _ teleportproto.RemoteExecutorServer = (*server)(nil)

func (s *server) Start(ctx context.Context, req *teleportproto.Command) (*teleportproto.JobStatus, error) {
	log.Println("Starting command", req.Command)
	job, err := s.jobs.Create(req.Command)
	if err != nil {
		return nil, errCouldNotStartProcess
	}
	return jobStatus(job.Status()), nil
}

func (s *server) Stop(ctx context.Context, req *teleportproto.JobId) (*teleportproto.JobStatus, error) {
	log.Println("Stopping job", req.Uuid)
	job, err := s.jobs.Stop(jobs.JobID(req.Uuid))
	if err != nil {
		if err == jobs.ErrNotFound {
			return nil, errIDNotFound
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
		return errIDNotFound
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
		return nil, errIDNotFound
	}
	return jobStatus(job.Status()), nil
}
