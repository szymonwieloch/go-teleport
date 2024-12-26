package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/szymonwieloch/go-teleport/client/proto/teleportproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func execute(args args) {
	if args.Start != nil {
		handleStart(args.Address, *args.Start)
	} else if args.Stop != nil {
		handleStop(args.Address, *args.Stop)
	} else if args.List != nil {
		handleList(args.Address, *args.List)
	} else if args.Log != nil {
		handleLog(args.Address, *args.Log)
	} else if args.Status != nil {
		handleStatus(args.Address, *args.Status)
	}

}

func createClient(addr string) (teleportproto.RemoteExecutorClient, func()) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	return teleportproto.NewRemoteExecutorClient(conn), func() { conn.Close() }
}

func handleStart(addr string, cmd startCmd) {
	log.Println("Starting command", cmd.Command)

	client, close := createClient(addr)
	defer close()
	ctx, cancel := defaultContext()
	defer cancel()
	req := teleportproto.Command{Command: []string{cmd.Command}}
	st, err := client.Start(ctx, &req)
	if err != nil {
		log.Fatalf("could not start a new command: %v", err)
	}
	log.Println("Started job", st.Id)
}

func handleStop(addr string, cmd stopCmd) {
	log.Println("Stopping job", cmd.JobID)
	client, close := createClient(addr)
	defer close()
	taskID := teleportproto.JobId{Uuid: string(cmd.JobID)}
	ctx, cancel := defaultContext()
	defer cancel()
	st, err := client.Stop(ctx, &taskID)
	if err != nil {
		log.Fatalf("could not stop the job: %v", err)
	}
	log.Println("Stopped job", cmd.JobID, "the error code was", st.ErrorCode)
}

func handleList(addr string, cmd listCmd) {
	log.Println("Listing jobs")
	client, close := createClient(addr)
	defer close()
	ctx, cancel := defaultContext()
	defer cancel()
	jobs, err := client.List(ctx, &empty.Empty{})
	if err != nil {
		log.Fatalf("could not list jobs: %v", err)
	}
	for _, job := range jobs.Jobs {
		log.Println(job.Id, job.JobStatus) // TODO: print more details
	}
}

func handleStatus(addr string, cmd statusCmd) {
	log.Println("Showing status for job", cmd.JobID)
	client, close := createClient(addr)
	defer close()
	ctx, cancel := defaultContext()
	defer cancel()
	taskID := teleportproto.JobId{Uuid: string(cmd.JobID)}
	status, err := client.GetStatus(ctx, &taskID)
	if err != nil {
		log.Fatalf("could not get status for the job: %v", err)
	}
	log.Println("Status for job", status.Id, "is", status.JobStatus)
}

func handleLog(addr string, cmd logCmd) {
	log.Println("Showing logs for job", cmd.JobID)
	client, close := createClient(addr)
	defer close()
	ctx, cancel := defaultContext()
	defer cancel()
	jobID := teleportproto.JobId{Uuid: string(cmd.JobID)}
	stream, err := client.Logs(ctx, &jobID)
	if err != nil {
		log.Fatalf("could not get logs for the job: %v", err)
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("== End of logs ==")
			return
		} else if err != nil {
			log.Fatalf("could not receive logs: %v", err)
		}
		if resp.Text != "" {
			switch resp.Src {
			case teleportproto.LogSource_LS_STDOUT:
				fmt.Print(resp.Text)

			case teleportproto.LogSource_LS_STDERR:
				fmt.Fprint(os.Stderr, resp.Text)
			}
		}
	}
}

func defaultContext() (context.Context, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	return ctx, cancel
}
