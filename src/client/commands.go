package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/szymonwieloch/go-teleport/client/proto/teleportproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const separator = "------------------------------------------------------------"

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
		fatalError(err, "did not connect to the server")
	}
	return teleportproto.NewRemoteExecutorClient(conn), func() { conn.Close() }
}

func handleStart(addr string, cmd startCmd) {
	fmt.Println("Starting command:", strings.Join(cmd.Command, " "))

	client, close := createClient(addr)
	defer close()
	ctx, cancel := defaultContext()
	defer cancel()
	req := teleportproto.Command{Command: cmd.Command}
	st, err := client.Start(ctx, &req)
	if err != nil {
		fatalError(err, "could not start a new command")
	}
	fmt.Println("Started job", st.Id.Uuid)
}

func handleStop(addr string, cmd stopCmd) {
	fmt.Println("Stopping job", cmd.JobID)
	client, close := createClient(addr)
	defer close()
	taskID := teleportproto.JobId{Uuid: string(cmd.JobID)}
	ctx, cancel := defaultContext()
	defer cancel()
	st, err := client.Stop(ctx, &taskID)
	if err != nil {
		fatalError(err, "could not stop the job")
	}
	fmt.Println("Stopped job")
	printStatus(st)
}

func handleList(addr string, cmd listCmd) {
	fmt.Println("Listing jobs")
	client, close := createClient(addr)
	defer close()
	ctx, cancel := defaultContext()
	defer cancel()
	list, err := client.List(ctx, &empty.Empty{})
	if err != nil {
		fatalError(err, "could not list jobs")
	}
	for _, status := range list.Jobs {
		fmt.Println(separator)
		printStatus(status)
	}
}

func handleStatus(addr string, cmd statusCmd) {
	fmt.Println("Showing status for job", cmd.JobID)
	client, close := createClient(addr)
	defer close()
	ctx, cancel := defaultContext()
	defer cancel()
	taskID := teleportproto.JobId{Uuid: string(cmd.JobID)}
	status, err := client.GetStatus(ctx, &taskID)
	if err != nil {
		fatalError(err, "could not get status for the job")
	}
	printStatus(status)
}

func handleLog(addr string, cmd logCmd) {
	fmt.Println("Showing logs for job", cmd.JobID)
	client, close := createClient(addr)
	defer close()
	ctx := context.Background() //defaultContext() cancel :=
	// defer cancel()
	jobID := teleportproto.JobId{Uuid: string(cmd.JobID)}
	stream, err := client.Logs(ctx, &jobID)
	if err != nil {
		fatalError(err, "could not get logs for the job")
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("== End of logs ==")
			return
		} else if err != nil {
			fatalError(err, "could not receive logs")
		}
		if resp.Text != "" {
			switch resp.Src {
			case teleportproto.LogSource_LS_STDOUT:
				fmt.Print(colorGreen + resp.Text + colorReset)

			case teleportproto.LogSource_LS_STDERR:
				fmt.Fprint(os.Stderr, colorRed+resp.Text)
			}
		}
	}
}

func defaultContext() (context.Context, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	return ctx, cancel
}

func printStatus(status *teleportproto.JobStatus) {
	fmt.Println("Job ID :", status.Id.Uuid)
	fmt.Println("Command:", strings.Join(status.Command.Command, " "))
	fmt.Println("Started:", status.Started.AsTime())
	fmt.Println("Logs   :", status.Logs)
	if status.Details != nil {
		switch details := status.Details.(type) {
		case *teleportproto.JobStatus_Stopped:
			fmt.Println("Stopped:", details.Stopped.Stopped.AsTime())
			fmt.Println("E. code:", details.Stopped.ErrorCode)
		case *teleportproto.JobStatus_Pending:
			fmt.Println("CPU %  :", details.Pending.CpuPerc)
		}
	}
}

func fatalError(err error, msg string) {
	fmt.Fprintf(os.Stderr, msg+": %v\n", err)
	os.Exit(1)
}
